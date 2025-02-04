package datasource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
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
	connCfg.Addr = c.host
	connCfg.Enabled = true
	connCfg.UseTls = false

	c.conn = chdb.NewDB(connCfg)
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

func (c *Clickhouse) InitDBSchema(ctx context.Context, pgDSN string) error {
	if c.conn == nil {
		return errors.New("not connected")
	}

	expectedDbState, err := os.ReadFile(
		filepath.Join(projectpath.RootPath, "/db-clickhouse/", "db.sql"),
	)
	if err != nil {
		return err
	}

	sql := string(expectedDbState)
	if pgDSN != "" {
		u, err := url.Parse(pgDSN)
		if err != nil {
			return fmt.Errorf("cannot parse dsn: %w", err)
		}

		pwd, _ := u.User.Password()

		old := `PostgreSQL('db:5432', 'postgres', 'projects', 'postgres', '[HIDDEN]')`
		replace := fmt.Sprintf(
			`PostgreSQL('db:5432', '%s', 'projects', '%s', '%s')`,
			u.Path[1:],
			u.User.Username(),
			pwd,
		)
		sql = strings.ReplaceAll(sql, old, replace)
	}

	return c.conn.ExecMultiContext(ctx, sql)
}

func (c *Clickhouse) createDB(ctx context.Context, db *chdb.DB, name string) error {
	return db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", name))
}

func (c *Clickhouse) dropDB(ctx context.Context, db *chdb.DB, name string) error {
	return db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s", name))
}
