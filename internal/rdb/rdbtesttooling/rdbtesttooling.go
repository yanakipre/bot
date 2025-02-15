package rdbtesttooling

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // postgres driver
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/postgres"
	"github.com/rekby/fixenv"
	"github.com/testcontainers/testcontainers-go"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/secret"
)

const DsnEnvVar = "TEST_DATABASE_URL"

// RealHostname changes host to host.docker.internal if ran inside docker (GNOMOCK_ENV=gnomockd)
// so it would be able to connect to port, opened on host machine (or forwarded to other container)
func RealHostname(host string) string {
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

	conn, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}

	return conn, conn.Ping()
}

func postgresHealthcheck(ctx context.Context, c *gnomock.Container) error {
	if c != nil {
		c.Host = RealHostname(c.Host)
	}
	db, err := postgresConnect(c, "postgres")
	if err != nil {
		return err
	}

	defer func() { _ = db.Close() }()

	var one int

	return db.QueryRowContext(ctx, `select 1`).Scan(&one)
}

// FixturePostgresProject creates a Postgres project of 1 host.
// It caches result per PACKAGE. This works ONLY if TestMain is configured correctly.
func FixturePostgresProject(e fixenv.Env) secret.String {
	return fixenv.CacheResult(e, func() (*fixenv.GenericResult[secret.String], error) {
		dsn, ok := os.LookupEnv(DsnEnvVar)
		if ok && dsn != "" {
			return fixenv.NewGenericResult(secret.NewString(dsn)), nil
		}
		container, err := StartPGContainer("14-alpine")
		if err != nil {
			panic(fmt.Errorf("could not start postgres project: %w", err))
		}
		cleanup := func() { _ = gnomock.Stop(container) }
		dsn = fmt.Sprintf("postgres://postgres:password@%s:%d/postgres",
			container.Host,
			container.DefaultPort(),
		)
		return fixenv.NewGenericResultWithCleanup(secret.NewString(dsn), cleanup), nil
	}, fixenv.CacheOptions{
		Scope:    fixenv.ScopePackage,
		CacheKey: "postgres-project",
	})
}

func StartReusablePGContainer(version, containerName string) (*gnomock.Container, error) {
	return StartPGContainerFromPreset(
		postgres.Preset(postgres.WithVersion(version)),
		gnomock.WithHealthCheck(postgresHealthcheck),
		gnomock.WithContainerReuse(),
		gnomock.WithContainerName(containerName),
	)
}

func StartPGContainer(version string) (*gnomock.Container, error) {
	return StartPGContainerFromPreset(
		postgres.Preset(postgres.WithVersion(version)),
		gnomock.WithCommand("postgres", "-c", "wal_level=logical"),
		gnomock.WithHealthCheck(postgresHealthcheck),
	)
}

func StartPGContainerFromPreset(p gnomock.Preset, opts ...gnomock.Option) (*gnomock.Container, error) {
	pgp := p.(*postgres.P)

	// reuse docker image cache without network round trips to docker hub.
	switch pgp.Version {
	case "14-alpine":
		pgp.Version = "sha256:51ce26e4463d434049b4b83e72eaaa008047a6a6cc65f2f3ee2ff3c183da0621"
	default:
		panic("unsupported version")
	}

	// These settings tell postgres acknowledge commits and make updates
	// before flushing records to disk
	// More details:
	// https://www.percona.com/blog/2020/08/21/postgresql-synchronous_commit-options-and-synchronous-standby-replication/
	// https://www.postgresql.org/docs/8.1/runtime-config-wal.html
	pgp.Queries = append([]string{
		"ALTER SYSTEM SET synchronous_commit=off;",
		"ALTER SYSTEM SET fsync=off;",
		"SELECT pg_reload_conf();",
	}, pgp.Queries...)

	p = &fixImage{P: pgp}

	c, err := gnomock.Start(p, append(opts, WithRegistryAuth(p.Image()))...)
	if c != nil {
		c.Host = RealHostname(c.Host)
	}
	return c, err
}

type fixImage struct {
	*postgres.P
}

func (p *fixImage) Image() string {
	if strings.Contains(p.Version, "sha256") {
		return fmt.Sprintf("docker.io/library/postgres@%s", p.Version)
	}
	return p.P.Image()
}

func WithRegistryAuth(image string) gnomock.Option {
	ctx := context.Background()

	// It seems that when we run the docker/login-action github action,
	// the registry name that appears in our docker config is https://index.docker.io/v1/.
	// If we pass docker.io/... to testcontainers.DockerImageAuth, it will fail to match it to the config.
	// But if we pass just the image name, it'll fall back to its default registry
	// and pick up the credentials from our docker config correctly.
	image = strings.TrimPrefix(image, "docker.io/library/")

	// copied and adapted from testcontainers attemptToPullImage()
	registry, imageAuth, err := testcontainers.DockerImageAuth(ctx, image)
	if err != nil {
		logger.Warn(ctx, fmt.Sprintf(
			"Failed to get image auth for %s. Setting empty credentials for the image: %s. Error is: %s",
			registry, image, err))
		return func(*gnomock.Options) {}
	}
	encodedJSON, err := json.Marshal(imageAuth)
	if err != nil {
		logger.Warn(ctx, fmt.Sprintf(
			"Failed to marshal image auth. Setting empty credentials for the image: %s. Error is: %s", image, err))
		return func(*gnomock.Options) {}
	} else {
		return gnomock.WithRegistryAuth(base64.URLEncoding.EncodeToString(encodedJSON))
	}
}
