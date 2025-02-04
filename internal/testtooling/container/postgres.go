package container

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
)

const (
	defaultPostgresHostname = "db"
)

type Postgres struct {
	container

	username string
	password string
	dbName   string
}

func NewPostgres(dbName, username, password string, opts ...Option) *Postgres {
	c := Postgres{
		username:  username,
		password:  password,
		dbName:    dbName,
		container: newContainer(defaultPostgresHostname),
	}
	c.sourceImage = "docker.io/postgres:14-alpine" // default

	for _, opt := range opts {
		opt(&c.container)
	}

	return &c
}

func NewPostgresWithDSN(dsn string, opts ...Option) (*Postgres, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot parse dsn: %w", err)
	}

	pwd, _ := u.User.Password()
	return NewPostgres(u.Path[1:], u.User.Username(), pwd, opts...), nil
}

func (c *Postgres) Run(ctx context.Context) error {
	pg, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage(c.sourceImage),
		postgres.WithUsername(c.username),
		postgres.WithPassword(c.password),
		postgres.WithDatabase(c.dbName),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Hostname: c.container.hostName,
				Networks: c.container.networks,
			},
		}),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return fmt.Errorf("could not start postgres container: %w", err)
	}

	// Container's ports debugging information.
	ports, _ := pg.Ports(ctx)
	logger.Debug(ctx, "postgres container's ports", zap.Any("ports", ports))

	addr, err := portEndpoint(ctx, pg, "5432/tcp", false)
	if err != nil {
		return fmt.Errorf("could not get postgres port: %w", err)
	}

	logger.Debug(ctx, "postgres container's exposed address", zap.String("address", addr.String()))

	c.container.container = pg
	c.container.exposedAddr = addr

	return nil
}
