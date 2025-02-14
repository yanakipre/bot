package testtooling

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/yanakipre/bot/internal/clouderr"
	"github.com/yanakipre/bot/internal/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"math"
	"os"
	"strings"
	"sync/atomic"

	"github.com/jackc/pgx/v4"
	"github.com/rekby/fixenv"

	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/rdb"
	"github.com/yanakipre/bot/internal/rdb/rdbtesttooling"
	"github.com/yanakipre/bot/internal/secret"
	"github.com/yanakipre/bot/internal/testtooling/container"
	"github.com/yanakipre/bot/internal/testtooling/datasource"
)

var testDBIdx uint32

func DbName(dbProject string, name string) string {
	idx := atomic.AddUint32(&testDBIdx, 1)
	name = strings.ToLower(fmt.Sprintf("db%d_%s_%s", idx, dbProject, name))

	var result strings.Builder

	// replace all the characters except for alfa-numeric ones.
	for i := 0; i < len(name); i++ {
		b := name[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') {
			result.WriteByte(b)
		} else {
			result.WriteByte(byte('_'))
		}
	}

	return result.String()[:int(math.Min(64, float64(len(name))))]
}

func SetupDBWithName(
	ctx context.Context,
	cfg rdb.Config,
	dbName string,
) (*rdb.DB, func() error, error) {
	connStrParts := strings.Split(cfg.DSN.Unmask(), "/")
	connStrParts[3] = dbName
	cfg.DSN = secret.NewString(strings.Join(connStrParts, "/"))

	connCfg, err := pgx.ParseConfig(cfg.DSN.Unmask())
	if err != nil {
		return nil, nil, err
	}

	connCfg.Database = dbName

	cfg.DSN = secret.NewValue(connCfg.ConnString())

	managementConnCfg := connCfg.Copy()
	err = dbPostgresCreate(ctx, managementConnCfg, dbName)
	if err != nil {
		var pgerr *pgconn.PgError
		if !errors.As(err, &pgerr) || pgerr.Code != pgerrcode.DuplicateDatabase {
			return nil, nil, err
		}
	}

	db := rdb.New(cfg)

	teardown := func() error {
		if err := db.Close(); err != nil {
			return err
		}
		if err := dbPostgresDropSubscriptions(ctx, connCfg, dbName); err != nil {
			return err
		}
		return dbPostgresDrop(ctx, managementConnCfg, dbName)
	}

	return db, teardown, db.Ready(ctx)
}

const managementPostgresDB = "postgres"

func dbPostgresDrop(ctx context.Context, cfg *pgx.ConnConfig, name string) error {
	cfg.Database = managementPostgresDB
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE %s", name))
	if err != nil {
		return err
	}
	if err = conn.Close(ctx); err != nil {
		return err
	}
	return nil
}

func dbPostgresDropSubscriptions(ctx context.Context, cfg *pgx.ConnConfig, name string) error {
	cfg.Database = name
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}
	rows, err := conn.Query(ctx, "SELECT subname FROM pg_subscription")
	if err != nil {
		return err
	}
	defer rows.Close()
	subscriptions := []string{}
	for rows.Next() {
		var subname string
		if err := rows.Scan(&subname); err != nil {
			return err
		}
		subscriptions = append(subscriptions, subname)
	}
	rows.Close()

	for _, subname := range subscriptions {
		_, err = conn.Exec(ctx, fmt.Sprintf("ALTER SUBSCRIPTION %s DISABLE", subname))
		if err != nil {
			if strings.HasSuffix(err.Error(), fmt.Sprintf(`subscription "%s" does not exist (SQLSTATE 42704)`, subname)) {
				logger.Warn(ctx, "subscription does not exist", zap.String("subscription", subname))
				continue
			}
			return err
		}
		_, err = conn.Exec(ctx, fmt.Sprintf("ALTER SUBSCRIPTION %s SET (slot_name = NONE)", subname))
		if err != nil {
			return err
		}
		_, err = conn.Exec(ctx, fmt.Sprintf("DROP SUBSCRIPTION %s", subname))
		if err != nil {
			return err
		}
	}
	if err = conn.Close(ctx); err != nil {
		return err
	}
	return nil
}

func dbPostgresCreate(ctx context.Context, cfg *pgx.ConnConfig, name string) error {
	cfg.Database = managementPostgresDB
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}
	if _, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", name)); err != nil {
		return err
	}
	if err = conn.Close(ctx); err != nil {
		return err
	}
	return nil
}

func FixtureReusableDBWithSchemaSeeded(e fixenv.Env, name, schemaPath, lastMigrationPath, seedPath string) *rdb.DB {
	return fixenv.CacheResult(e, func() (*fixenv.GenericResult[*rdb.DB], error) {
		c, err := rdbtesttooling.StartReusablePGContainer("14-alpine", "reusable-test-db")
		if err != nil {
			panic(fmt.Errorf("could not start postgres project: %w", err))
		}

		hash, err := getLastMigrationHash(lastMigrationPath)
		if err != nil {
			panic(fmt.Errorf("could not get last migration hash: %w", err))
		}
		// We're using a name and the last migration hash as the database name.
		// The hash is included, so that when a developer checks out a branch with another schema version
		// and runs tests, we'll simply create a new schema instead of failing and keeping the old schema.
		dbName := name + "_" + hash

		cfg := rdb.DefaultConfig()
		cfg.DSN = secret.NewString(fmt.Sprintf("postgres://postgres:password@%s:%d/postgres", c.Host, c.DefaultPort()))
		cfg.SearchPath = "public"
		db, _, err := SetupDBWithName(context.Background(), cfg, dbName)
		if err != nil {
			panic(fmt.Sprintf("could not setup database in postgres project: %v", err))
		}

		db.MapperFunc(encodingtooling.CamelToSnake)

		if err := ExecSQLFile(db, schemaPath); err != nil {
			var pgerr *pgconn.PgError
			if !errors.As(err, &pgerr) ||
				(pgerr.Code != pgerrcode.DuplicateSchema &&
					pgerr.Code != pgerrcode.DuplicateObject &&
					pgerr.Code != pgerrcode.DuplicateFunction &&
					pgerr.Code != pgerrcode.DuplicateTable) {
				panic(fmt.Sprintf("could not create db with schema: %v", err))
			}
		} else {
			// Only seed data if there was no error.
			// Otherwise, we assume that the schema and seed data already exist.
			if err := ExecSQLFile(db, seedPath); err != nil {
				var pgerr *pgconn.PgError
				if !errors.As(err, &pgerr) || pgerr.Code != pgerrcode.UniqueViolation {
					panic(fmt.Sprintf("could not seed db: %v", err))
				}
			}
		}

		return fixenv.NewGenericResult(db), nil
	}, fixenv.CacheOptions{
		Scope:    fixenv.ScopeTest,
		CacheKey: "postgres-seeded-" + name,
	})
}

func getLastMigrationHash(path string) (string, error) {
	file, err := os.Open(path) //nolint:gosec
	if err != nil {
		return "", err
	}
	var last struct {
		Hash string `json:"hash"`
	}
	if err := json.NewDecoder(file).Decode(&last); err != nil {
		return "", err
	}
	return strings.ToLower(last.Hash), nil
}

// FixtureDBWithSchema creates database and schema every non cached call,
// caches result for each call in each test.
// and destroys db after each test.
func FixtureDBWithSchema(e fixenv.Env, schemaPath string) *rdb.DB {
	return fixenv.CacheResult(e, func() (*fixenv.GenericResult[*rdb.DB], error) {
		db := FixtureEmptyPostgresDB(e, "public")
		err := ExecSQLFile(db, schemaPath)
		if err != nil {
			panic(fmt.Errorf("could not create db with schema: %w", err))
		}
		return fixenv.NewGenericResult(db), nil
	}, fixenv.CacheOptions{
		Scope:    fixenv.ScopeTest,
		CacheKey: "postgres-schema-" + schemaPath,
	})
}

func ExecSQLFile(db *rdb.DB, path string) error {
	sql, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return err
	}
	// Use a noop logger. We expect certain errors (like schema already exists) and don't want to spam test output.
	ctx := logger.WithLogger(context.Background(), logger.New(zapcore.FatalLevel, &bytes.Buffer{}, nil))
	_, err = db.ExecContext(ctx, string(sql))
	return err
}

// FixtureEmptyPostgresDB creates database every non cached call,
// caches result for each call in each test.
// and destroys db after each test.
func FixtureEmptyPostgresDB(e fixenv.Env, dbProject string) *rdb.DB {
	return fixenv.CacheResult(e, func() (*fixenv.GenericResult[*rdb.DB], error) {
		dsn := rdbtesttooling.FixturePostgresProject(e)
		t := e.T()
		randName := DbName(dbProject, t.Name())
		ctx := context.Background()
		cfg := rdb.DefaultConfig()
		cfg.DSN = dsn
		cfg.SearchPath = "public"
		db, td, err := SetupDBWithName(ctx, cfg, randName)
		if err != nil {
			panic(fmt.Errorf(
				"could not setup database in postgres project: %w",
				clouderr.WrapWithFields(err, zap.String("database_name", randName))),
			)
		}
		cleanup := func() {
			if err := td(); err != nil {
				panic(fmt.Errorf("error while destroying postgres project: %w", err))
			}
		}
		db.MapperFunc(encodingtooling.CamelToSnake)
		return fixenv.NewGenericResultWithCleanup(db, cleanup), nil
	}, fixenv.CacheOptions{
		Scope:    fixenv.ScopeTest,
		CacheKey: fmt.Sprintf("postgres-db-%s", dbProject),
	})
}

func FixturePostgresContainer(ctx context.Context, e fixenv.Env) *container.Postgres {
	return fixenv.CacheResult(e, func() (*fixenv.GenericResult[*container.Postgres], error) {
		pgParams := rdb.DefaultConfig()

		net := ContainerNetwork(ctx, e)

		ch, err := container.NewPostgresWithDSN(pgParams.DSN.Unmask(), container.WithNetworks(net.Name()))
		if err != nil {
			e.T().Fatalf("cannot initialize postgres container: %v", err)
			return nil, err
		}

		if err = ch.Run(ctx); err != nil {
			e.T().Fatalf("cannot run postgres container: %v", err)
			return nil, err
		}

		cleanup := func() {
			if err = ch.Stop(ctx); err != nil {
				e.T().Fatalf("cannot stop postgres container: %v", err)
			}
		}

		return fixenv.NewGenericResultWithCleanup(ch, cleanup), nil
	}, fixenv.CacheOptions{
		Scope:    fixenv.ScopePackage,
		CacheKey: "postgres-container",
	})
}

func FixtureEmptyPostgresSchema(ctx context.Context, e fixenv.Env) *rdb.DB {
	return fixenv.CacheResult(e, func() (*fixenv.GenericResult[*rdb.DB], error) {
		t := e.T()
		randName := DbName("", t.Name())

		pgContainer := FixturePostgresContainer(ctx, e)

		pg := datasource.NewPostgresConnection(pgContainer.ExposedAddr().StringWithoutProto())
		if err := pg.CreateDatabase(ctx, randName); err != nil {
			e.T().Fatalf("cannot create a new Postgres DB: %v", err)
			return nil, err
		}

		pgConn, err := pg.Connect(ctx, randName)
		if err != nil {
			e.T().Fatalf("cannot connect to Postgres DB: %v", err)
			return nil, err
		}

		if err = pg.InitDBSchema(ctx); err != nil {
			e.T().Fatalf("cannot connect initialize Postgres schema: %v", err)
			return nil, err
		}

		cleanup := func() {
			if err = pg.Close(); err != nil {
				e.T().Fatalf("cannot close Postgres connection: %v", err)
				return
			}

			if err = pg.DropDatabase(ctx, randName); err != nil {
				panic(fmt.Errorf("error while destroying postgres project: %w", err))
			}
		}

		return fixenv.NewGenericResultWithCleanup(pgConn, cleanup), nil
	}, fixenv.CacheOptions{
		Scope:    fixenv.ScopeTest,
		CacheKey: "postgres-db",
	})
}
