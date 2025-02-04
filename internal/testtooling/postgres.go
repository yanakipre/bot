package testtooling

import (
	"context"
	"fmt"
	"math"
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
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, "'", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "#", "_")
	name = strings.ReplaceAll(name, ",", "")
	name = strings.ToLower(name)
	return name[0:int(math.Min(64, float64(len(name))))]
}

func SetupDBWithName(
	ctx context.Context,
	cfg rdb.Config,
	DBName string,
) (*rdb.DB, func() error, error) {
	connStrParts := strings.Split(cfg.DSN.Unmask(), "/")
	connStrParts[3] = DBName
	cfg.DSN = secret.NewString(strings.Join(connStrParts, "/"))

	connCfg, err := pgx.ParseConfig(cfg.DSN.Unmask())
	if err != nil {
		return nil, nil, err
	}

	connCfg.Database = DBName

	cfg.DSN = secret.NewValue(connCfg.ConnString())

	managementConnCfg := connCfg.Copy()
	err = dbPostgresCreate(ctx, managementConnCfg, DBName)
	if err != nil {
		return nil, nil, err
	}

	db := rdb.New(cfg)

	teardown := func() error {
		err := db.Close()
		if err != nil {
			return err
		}

		return dbPostgresDrop(ctx, managementConnCfg, DBName)
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

// FixtureEmptyPostgresDB creates database every non cached call,
// caches result for each call in each test.
// and destroys db after each test.
func FixtureEmptyPostgresDB(e fixenv.Env, dbProject string) *rdb.DB {
	cacheKey := fmt.Sprintf("postgres-db-%s", dbProject)
	return e.CacheWithCleanup(cacheKey, &fixenv.FixtureOptions{
		Scope: fixenv.ScopeTest,
	}, func() (res any, cleanup fixenv.FixtureCleanupFunc, err error) {
		dsn := rdbtesttooling.FixturePostgresProject(e)
		t := e.T()
		randName := DbName(dbProject, t.Name())
		ctx := context.Background()
		cfg := rdb.DefaultConfig()
		cfg.DSN = dsn
		cfg.SearchPath = "public"
		db, td, err := SetupDBWithName(ctx, cfg, randName)
		if err != nil {
			panic(fmt.Sprintf("could not setup database %q in postgres project %v", randName, err))
		}
		cleanup = func() {
			if err := td(); err != nil {
				panic(fmt.Sprintf("error while destroying postgres project: %s", err))
			}
		}
		db.MapperFunc(encodingtooling.CamelToSnake)
		return db, cleanup, nil
	}).(*rdb.DB)
}

func FixturePostgresContainer(ctx context.Context, e fixenv.Env) *container.Postgres {
	cacheKey := "postgres-container"
	return e.CacheWithCleanup(cacheKey, &fixenv.FixtureOptions{
		Scope: fixenv.ScopePackage,
	}, func() (any, fixenv.FixtureCleanupFunc, error) {
		pgParams := rdb.DefaultConfig()

		net := ContainerNetwork(ctx, e)

		ch, err := container.NewPostgresWithDSN(pgParams.DSN.Unmask(), container.WithNetworks(net.Name()))
		if err != nil {
			e.T().Fatalf("cannot initialize postgres container: %v", err)
			return nil, nil, err
		}

		if err = ch.Run(ctx); err != nil {
			e.T().Fatalf("cannot run postgres container: %v", err)
			return nil, nil, err
		}

		cleanup := func() {
			if err = ch.Stop(ctx); err != nil {
				e.T().Fatalf("cannot stop postgres container: %v", err)
			}
		}

		return ch, cleanup, nil
	}).(*container.Postgres)
}

func FixtureEmptyPostgresSchema(ctx context.Context, e fixenv.Env) *rdb.DB {
	cacheKey := "postgres-db"
	return e.CacheWithCleanup(cacheKey, &fixenv.FixtureOptions{
		Scope: fixenv.ScopeTest,
	}, func() (any, fixenv.FixtureCleanupFunc, error) {
		t := e.T()
		randName := DbName("", t.Name())

		pgContainer := FixturePostgresContainer(ctx, e)

		pg := datasource.NewPostgresConnection(pgContainer.ExposedAddr().StringWithoutProto())
		if err := pg.CreateDatabase(ctx, randName); err != nil {
			e.T().Fatalf("cannot create a new Postgres DB: %v", err)
			return nil, nil, err
		}

		pgConn, err := pg.Connect(ctx, randName)
		if err != nil {
			e.T().Fatalf("cannot connect to Postgres DB: %v", err)
			return nil, nil, err
		}

		if err = pg.InitDBSchema(ctx); err != nil {
			e.T().Fatalf("cannot connect initialize Postgres schema: %v", err)
			return nil, nil, err
		}

		cleanup := func() {
			if err = pg.Close(); err != nil {
				e.T().Fatalf("cannot close Postgres connection: %v", err)
				return
			}

			if err = pg.DropDatabase(ctx, randName); err != nil {
				panic(fmt.Sprintf("error while destroying postgres project: %s", err))
			}
		}

		return pgConn, cleanup, nil
	}).(*rdb.DB)
}
