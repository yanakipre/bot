package datasource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/yanakipre/bot/internal/projectpath"
	"github.com/yanakipre/bot/internal/rdb"
	"github.com/yanakipre/bot/internal/secret"
)

const (
	defaultManagementPostgresDB = "postgres"
	defaultSearchPath           = "public"
)

type Postgres struct {
	managementDBName string

	host string
	conn *rdb.DB
}

func NewPostgresConnection(host string) *Postgres {
	c := Postgres{
		host:             host,
		managementDBName: defaultManagementPostgresDB,
	}

	return &c
}

func (c *Postgres) Connect(ctx context.Context, dbName string) (*rdb.DB, error) {
	c.conn = rdb.New(c.defaultConfigWithDB(dbName))
	if err := c.conn.Ready(ctx); err != nil {
		return nil, fmt.Errorf("db is not ready: %w", err)
	}

	return c.conn, nil
}

func (c *Postgres) CreateDatabase(ctx context.Context, dbName string) error {
	managementDB := rdb.New(c.defaultConfigWithDB(c.managementDBName))
	if err := managementDB.Ready(ctx); err != nil {
		return fmt.Errorf("db is not ready: %w", err)
	}

	if err := c.createDB(ctx, managementDB, dbName); err != nil {
		return fmt.Errorf("cannot create db: %w", err)
	}

	return managementDB.Close()
}

func (c *Postgres) DropDatabase(ctx context.Context, dbName string) error {
	managementDB := rdb.New(c.defaultConfigWithDB(c.managementDBName))
	if err := managementDB.Ready(ctx); err != nil {
		return fmt.Errorf("db is not ready: %w", err)
	}

	if err := c.dropDB(ctx, managementDB, dbName); err != nil {
		return fmt.Errorf("cannot drop db: %w", err)
	}

	return managementDB.Close()
}

func (c *Postgres) Close() error {
	return c.conn.Close()
}

func (c *Postgres) InitDBSchema(ctx context.Context) error {
	if c.conn == nil {
		return errors.New("not connected")
	}

	expectedDbState, err := os.ReadFile(filepath.Join(projectpath.RootPath, "/db/", "db.sql"))
	if err != nil {
		return err
	}

	_, err = c.conn.ExecContext(ctx, string(expectedDbState))

	return err
}

func (c *Postgres) createDB(ctx context.Context, db *rdb.DB, name string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", name))
	return err
}

func (c *Postgres) dropDB(ctx context.Context, db *rdb.DB, name string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s", name))
	return err
}

func (c *Postgres) defaultConfigWithDB(dbName string) rdb.Config {
	cfg := rdb.DefaultConfig()
	connURL, err := url.Parse(cfg.DSN.Unmask())
	if err != nil {
		panic(err)
	}

	user := connURL.User.Username()
	if pwd, found := connURL.User.Password(); found {
		user += ":" + pwd
	}

	cfg.DSN = secret.NewValue(fmt.Sprintf("postgres://%s@%s/%s",
		user,
		c.host,
		dbName,
	))

	cfg.SearchPath = defaultSearchPath

	return cfg
}
