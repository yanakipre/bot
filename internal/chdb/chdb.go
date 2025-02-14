package chdb

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/yanakipre/bot/internal/secret"
	"regexp"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
)

const (
	chcConnName = "chc"
)

type DB struct {
	cfg      Config
	conn     driver.Conn
	readOnly bool
}

func NewDB(cfg Config, readOnly bool) *DB {
	return &DB{
		cfg:      cfg,
		readOnly: readOnly,
	}
}

func chOptions(cfg Config, addr string, password secret.String) *clickhouse.Options {
	options := &clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: password.Unmask(),
		},
		Settings: clickhouse.Settings{
			"max_execution_time": cfg.MaxExecutionTime,
			"max_query_size":     cfg.MaxQuerySize,
			"max_result_rows":    cfg.MaxResultRows,
			"max_result_bytes":   cfg.MaxResultBytes,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:      time.Duration(10) * time.Second,
		MaxOpenConns:     cfg.MaxOpenConns,
		MaxIdleConns:     cfg.MaxIdleConns,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	}

	if cfg.UseTls {
		// DoubleCloud accepts only secure connections (like --secure in clickhouse-client)
		// but locally we don't have certificates, so skip TLS verification
		options.TLS = &tls.Config{} //nolint:gosec
	}

	return options
}

func (db *DB) Ready(ctx context.Context) error {
	if !db.cfg.Enabled {
		return nil
	}

	addr := db.cfg.ChcAddr
	if db.readOnly && db.cfg.ChcRoAddr != "" {
		addr = db.cfg.ChcRoAddr
	}

	options := chOptions(db.cfg, addr, db.cfg.ChcPassword)
	conn, err := clickhouse.Open(options)
	if err != nil {
		return fmt.Errorf("could not connect to clickhouse: %w", err)
	}
	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("could not ping clickhouse: %w", err)
	}
	db.conn = conn

	// Send database metrics
	go func() {
		timer := time.NewTicker(db.cfg.CollectMetricsInterval.Duration)
		defer timer.Stop()
		select {
		case <-timer.C:
			if db.conn != nil {
				sendMetrics(db.getConnName(), db.conn.Stats())
			}
		case <-ctx.Done():
			return
		}
	}()

	return nil
}

func (db *DB) getConnName() string {
	if db.readOnly {
		return chcConnName + "_ro"
	}

	return chcConnName
}

func (db *DB) IsReady() bool {
	if !db.cfg.Enabled {
		return false
	}

	return db.conn != nil
}

func (db *DB) Close() error {
	if !db.cfg.Enabled {
		return nil
	}

	if db.conn != nil {
		if err := db.conn.Close(); err != nil {
			return err
		}
	}

	return nil
}

func makeLoggerHook(ctx context.Context, connName, query string, args ...any) func() {
	now := time.Now()
	if len(args) > 1000 {
		args = args[:1000]
	}
	return func() {
		logger.Debug(ctx, "chdb.sql:query",
			zap.String("conn_name", connName),
			zap.String("query", query),
			logger.RawJSON("sql_args", args),
			zap.Duration("duration", time.Since(now)),
		)
	}
}

// TODO: bring logger hook as in rdb package https://github.com/yanakipredatabase/cloud/issues/7382
func (db *DB) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	if db.conn == nil {
		return ErrConnectionNotInitialized
	}

	defer makeLoggerHook(ctx, db.getConnName(), query, args)()
	return db.conn.Select(ctx, dest, query, args...)
}

func (db *DB) PrepareBatchContext(ctx context.Context, query string) (driver.Batch, error) {
	if db.conn == nil {
		return nil, ErrConnectionNotInitialized
	}

	if db.readOnly {
		return nil, ErrNotAvailableInRoMode
	}

	return db.conn.PrepareBatch(ctx, query)
}

func (db *DB) InsertContext(ctx context.Context, query string, args ...any) error {
	if db.conn == nil {
		return ErrConnectionNotInitialized
	}

	if db.readOnly {
		return ErrNotAvailableInRoMode
	}

	defer makeLoggerHook(ctx, db.getConnName(), query, args)()
	batch, err := db.PrepareBatchContext(ctx, query)
	if err != nil {
		return err
	}
	if err = batch.Append(args...); err != nil {
		return err
	}
	if err = batch.Send(); err != nil {
		return err
	}

	return nil
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) error {
	if db.conn == nil {
		return ErrConnectionNotInitialized
	}

	if db.readOnly {
		return ErrNotAvailableInRoMode
	}

	defer makeLoggerHook(ctx, db.getConnName(), query, args)()
	return db.conn.Exec(ctx, query, args...)
}

// ExecMultiContext executes query with multiple statements.
// clickhouse-go client doesn't support it automatically,
// so do as golang-migrate - split query by `;` and execute each statement.
func (db *DB) ExecMultiContext(ctx context.Context, query string) error {
	if db.conn == nil {
		return ErrConnectionNotInitialized
	}

	if db.readOnly {
		return ErrNotAvailableInRoMode
	}

	for _, stmt := range splitQuery(query) {
		if err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func splitQuery(query string) []string {
	var res []string

	removeCommentsReg := regexp.MustCompile(`--.*$`)
	for _, stmt := range strings.Split(query, ";") {
		stmt = strings.TrimSpace(stmt)
		stmt = removeCommentsReg.ReplaceAllString(stmt, "")
		if stmt == "" {
			continue
		}
		res = append(res, stmt)
	}
	return res
}
