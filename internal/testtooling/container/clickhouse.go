package container

import (
	"context"
	"fmt"
	"path"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
	"github.com/yanakipe/bot/internal/projectpath"
)

const (
	defaultClickhouseHostname = "clickhouse"
)

type Clickhouse struct {
	container

	username string
	password string
	dbName   string
}

func NewClickhouse(dbName, username, password string, opts ...Option) *Clickhouse {
	c := Clickhouse{
		username:  username,
		password:  password,
		dbName:    dbName,
		container: newContainer(defaultClickhouseHostname),
	}

	for _, opt := range opts {
		opt(&c.container)
	}

	return &c
}

func (c *Clickhouse) Run(ctx context.Context) error {
	ch, err := clickhouse.RunContainer(
		ctx,
		testcontainers.WithImage("clickhouse/clickhouse-server:23.8"),
		clickhouse.WithUsername(c.username),
		clickhouse.WithPassword(c.password),
		clickhouse.WithDatabase(c.dbName),
		clickhouse.WithConfigFile(
			path.Join(projectpath.RootPath, "/db-clickhouse/configs/test_config.xml"),
		),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Hostname: c.container.hostName,
				Networks: c.container.networks,
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("could not start clickhouse container: %w", err)
	}

	// Container's ports debugging information.
	ports, _ := ch.Ports(ctx)
	logger.Debug(ctx, "clickhouse container's ports", zap.Any("ports", ports))

	addr, err := portEndpoint(ctx, ch, "9000/tcp", false)
	if err != nil {
		return fmt.Errorf("could not get clickhouse port: %w", err)
	}

	logger.Debug(
		ctx,
		"clickhouse container's exposed address",
		zap.String("address", addr.String()),
	)

	c.container.container = ch
	c.container.exposedAddr = addr

	return nil
}
