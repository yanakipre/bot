package rdbtesttooling

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/orlangure/gnomock"
	"github.com/rekby/fixenv"
	"github.com/yanakipre/bot/internal/secret"
	"github.com/yanakipre/bot/internal/testtooling/container"
)

const DsnEnvVar = "TEST_DATABASE_URL"

// realHostname changes host to host.docker.internal if ran inside docker (GNOMOCK_ENV=gnomockd)
// so it would be able to connect to port, opened on host machine (or forwarded to other container)
func realHostname(host string) string {
	gnomockEnv := os.Getenv("GNOMOCK_ENV")
	if gnomockEnv == "gnomockd" {
		// we could patch it with arbitrary host, but inside gnomock
		// there is already GNOMOCK_ENV is used, so we will follow the same approach
		return "host.docker.internal"
	}
	return host
}

func postgresConnect(c *gnomock.Container, db string) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port(gnomock.DefaultPort),
		"postgres", "password", db, "disable",
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return conn, conn.Ping()
}

func postgresHealthcheck(ctx context.Context, c *gnomock.Container) error {
	if c != nil {
		c.Host = realHostname(c.Host)
	}
	db, err := postgresConnect(c, "postgres")
	if err != nil {
		return err
	}

	defer func() { _ = db.Close() }()

	var one int

	return db.QueryRow(`select 1`).Scan(&one)
}

func StartPGContainer(ctx context.Context, options ...container.Option) (*container.Postgres, error) {
	c := container.NewPostgres("postgres", "postgres", "password", options...)
	err := c.Run(ctx)
	if err != nil {
		return nil, err
	}

	// These settings tell postgres acknowledge commits and make updates
	// before flushing records to disk
	// More details:
	// https://www.percona.com/blog/2020/08/21/postgresql-synchronous_commit-options-and-synchronous-standby-replication/
	// https://www.postgresql.org/docs/8.1/runtime-config-wal.html
	// TODO: finish this
	//postgres.WithQueries(
	//	"ALTER SYSTEM SET synchronous_commit=off;",
	//	"ALTER SYSTEM SET fsync=off;",
	//	"SELECT pg_reload_conf();",
	//)

	return c, err
}

// FixturePostgresProject creates a Postgres project of 1 host.
// It caches result per PACKAGE. This works ONLY if TestMain is configured correctly.
func FixturePostgresProject(e fixenv.Env) secret.String {
	cacheKey := "postgres-project"
	return e.CacheWithCleanup(cacheKey, &fixenv.FixtureOptions{
		Scope: fixenv.ScopePackage,
	}, func() (res any, cleanup fixenv.FixtureCleanupFunc, err error) {
		dsn, ok := os.LookupEnv(DsnEnvVar)
		if ok && dsn != "" {
			return dsn, nil, nil
		}
		container, err := StartPGContainer(context.TODO())
		if err != nil {
			panic(fmt.Sprintf("could not start postgres project: %v", err))
		}
		cleanup = func() { _ = container.Stop(context.TODO()) }
		dsn = fmt.Sprintf("postgres://postgres:password@%s/postgres",
			container.ExposedAddr().StringWithoutProto(),
		)
		return secret.NewString(dsn), cleanup, nil
	}).(secret.String)
}
