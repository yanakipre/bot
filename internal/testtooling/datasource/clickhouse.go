package datasource

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yanakipre/bot/internal/clouderr"
	"go.uber.org/zap"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/yanakipre/bot/internal/chdb"
	"github.com/yanakipre/bot/internal/projectpath"
)

const ManagementClickHouseDB = "default"

type Clickhouse struct {
	managementDBName string

	host string
	conn *chdb.DB
}

func NewClickHouse(host string) *Clickhouse {
	c := Clickhouse{
		host:             host,
		managementDBName: ManagementClickHouseDB,
	}

	return &c
}

func (c *Clickhouse) Connect(ctx context.Context, dbName string) (*chdb.DB, error) {
	connCfg := chdb.DefaultConfig()
	connCfg.Database = dbName
	connCfg.ChcAddr = c.host
	connCfg.Enabled = true
	connCfg.UseTls = false

	c.conn = chdb.NewDB(connCfg, false)
	if err := c.conn.Ready(ctx); err != nil {
		return nil, fmt.Errorf("db is not ready: %w", err)
	}

	return c.conn, nil
}

func (c *Clickhouse) CreateDatabase(ctx context.Context, dbName string) error {
	conn, err := c.Connect(ctx, c.managementDBName)
	if err != nil {
		return err
	}

	c.conn = nil
	if err = c.createDB(ctx, conn, dbName); err != nil {
		return fmt.Errorf("cannot create db: %w", err)
	}

	return conn.Close()
}

func (c *Clickhouse) DropDatabase(ctx context.Context, dbName string) error {
	conn, err := c.Connect(ctx, c.managementDBName)
	if err != nil {
		return err
	}

	c.conn = nil
	if err = c.dropDB(ctx, conn, dbName); err != nil {
		return fmt.Errorf("cannot drop db: %w", err)
	}

	return conn.Close()
}

func (c *Clickhouse) Close() error {
	return c.conn.Close()
}

func (c *Clickhouse) InitDBSchema(ctx context.Context, pgDSN string, dbName string) error {
	if c.conn == nil {
		return errors.New("not connected")
	}

	chPort := strings.Split(c.host, ":")[1]

	if pgDSN == "" {
		pgDSN = "postgres://postgres:password@db:5432/yanakipre"
	}

	u, err := url.Parse(pgDSN)
	if err != nil {
		return fmt.Errorf("cannot parse dsn: %w", err)
	}
	pwd, _ := u.User.Password()

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    filepath.Join(projectpath.RootPath, "/db-clickhouse/"),
			Dockerfile: "migrations.Dockerfile",
			KeepImage:  true,
			Tag:        "yanakipre-ch-migration",
		},
		WaitingFor: wait.ForExit(),
		Cmd: []string{
			"--clickhouse-url",
			fmt.Sprintf(
				"clickhouse://host.docker.internal:%s?database=%s&x-multi-statement=true",
				chPort,
				dbName,
			),
			"--console-postgres-conn",
			fmt.Sprintf(
				"postgres://%s:%s@db:5432/%s",
				u.User.Username(),
				pwd,
				u.Path[1:],
			),
		},
		HostConfigModifier: func(config *container.HostConfig) {
			config.ExtraHosts = []string{"host.docker.internal:host-gateway"}
		},
	}

	ct, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return err
	}

	defer func() {
		_ = ct.Terminate(ctx)
	}()

	state, err := ct.State(ctx)
	if err != nil {
		return err
	}

	if state.ExitCode != 0 {
		logs, err := ct.Logs(ctx)
		if err != nil {
			return err
		}

		logsString, err := io.ReadAll(logs)
		if err != nil {
			return err
		}

		return clouderr.WithFields(
			"container exited with a non-zero exit code",
			zap.String("logs", string(logsString)),
			zap.Int("exit_code", state.ExitCode),
		)
	}

	return nil
}

func (c *Clickhouse) createDB(ctx context.Context, db *chdb.DB, name string) error {
	return db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", name))
}

func (c *Clickhouse) dropDB(ctx context.Context, db *chdb.DB, name string) error {
	return db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s", name))
}
