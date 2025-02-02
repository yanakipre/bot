package driver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"go.uber.org/zap/zapcore"

	"github.com/yanakipe/bot/internal/rdb/internal/driver/internal/hooks"
	"github.com/yanakipe/bot/internal/rdb/internal/driver/internal/wrapper"
)

type openOptions struct {
	searchPath   wrapper.AfterConnQuery
	dbname       string
	hostname     string
	logging      bool
	loggingLevel zapcore.Level
	hooks        []wrapper.Hook
}

type OpenOption func(options *openOptions)

func WithDbName(dbname string) OpenOption {
	return func(options *openOptions) {
		options.dbname = dbname
	}
}

// WithSearchPath sets search path after connection has been established.
func WithSearchPath(path string) OpenOption {
	query := fmt.Sprintf("SET search_path TO %s", path)
	return func(options *openOptions) {
		options.searchPath = func(ctx context.Context, conn driver.Conn) error {
			if c, ok := conn.(interface {
				ExecContext(
					ctx context.Context,
					query string,
					args []driver.NamedValue,
				) (res driver.Result, err error)
			}); ok {
				_, err := c.ExecContext(ctx, query, nil)
				return err
			}
			panic("cannot apply search path - connection does not support ExecContext")
		}
	}
}

func WithHostname(hostname string) OpenOption {
	return func(options *openOptions) {
		options.hostname = hostname
	}
}

func WithLogging(lvl zapcore.Level) OpenOption {
	return func(options *openOptions) {
		options.logging = true
		options.loggingLevel = lvl
	}
}

func WithDSNParsedInfo(dsn string) OpenOption {
	return func(options *openOptions) {
		options.hostname, options.dbname, _ = parseDSN(dsn)
	}
}

func Open(driverName string, dsn string, opts ...OpenOption) (*sql.DB, error) {
	host, dbname, _ := parseDSN(dsn)

	options := openOptions{
		loggingLevel: zapcore.DebugLevel,
		hostname:     host,
		dbname:       dbname,
	}
	for _, opt := range opts {
		opt(&options)
	}

	var wrapperHooks []wrapper.Hook
	if options.logging {
		wrapperHooks = append(
			wrapperHooks,
			hooks.NewLogger(options.loggingLevel, options.hostname, options.dbname),
		)
	}
	wrapperHooks = append(wrapperHooks, options.hooks...)

	var wrapperHook wrapper.Hook
	if len(wrapperHooks) == 1 {
		wrapperHook = wrapperHooks[0]
	} else {
		wrapperHook = hooks.Compose(wrapperHooks...)
	}

	wrappedDriverName, err := wrapper.WrapDriverByName(driverName, wrapperHook, options.searchPath)
	if err != nil {
		return nil, err
	}
	dn := wrappedDriverName
	db, err := sql.Open(dn, dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// UnwrapConn return parent driver.Conn
func UnwrapConn(c driver.Conn) driver.Conn {
	return wrapper.UnwrapConn(c)
}
