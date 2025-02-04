package chdb

import (
	"context"
	"crypto/tls"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
)

type DB struct {
	cfg  Config
	conn driver.Conn
}

func NewDB(cfg Config) *DB {
	return &DB{
		cfg: cfg,
	}
}

func (db *DB) Ready(ctx context.Context) error {
	if !db.cfg.Enabled {
		return nil
	}

	options := &clickhouse.Options{
		Addr: []string{db.cfg.Addr},
		Auth: clickhouse.Auth{
			Database: db.cfg.Database,
			Username: db.cfg.Username,
			Password: db.cfg.Password.Unmask(),
		},
		Settings: clickhouse.Settings{
			"max_execution_time": db.cfg.MaxExecutionTime,
			"max_query_size":     db.cfg.MaxQuerySize,
			"max_result_rows":    db.cfg.MaxResultRows,
			"max_result_bytes":   db.cfg.MaxResultBytes,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:      time.Duration(10) * time.Second,
		MaxOpenConns:     db.cfg.MaxOpenConns,
		MaxIdleConns:     db.cfg.MaxIdleConns,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	}
	if db.cfg.UseTls {
		// DoubleCloud accepts only secure connections (like --secure in clickhouse-client)
		// but locally we don't have certificates, so skip TLS verification
		options.TLS = &tls.Config{} //nolint:gosec
	}
	conn, err := clickhouse.Open(options)
	if err != nil {
		return fmt.Errorf("could not connec to to ClickHouse cluster: %w", err)
	}
	db.conn = conn

	// Send database metrics
	go func() {
		timer := time.NewTicker(db.cfg.CollectMetricsInterval.Duration)
		defer timer.Stop()
		select {
		case <-timer.C:
			sendMetrics(db.conn.Stats())
		case <-ctx.Done():
			return
		}
	}()

	return conn.Ping(ctx)
}

func (db *DB) Close() error {
	if !db.cfg.Enabled {
		return nil
	}
	return db.conn.Close()
}

func makeLoggerHook(ctx context.Context, query string, args ...any) func() {
	now := time.Now()
	if len(args) > 1000 {
		args = args[:1000]
	}
	return func() {
		logger.Debug(ctx, "chdb.sql:query",
			zap.String("sql_statement", query),
			zap.Any("sql_args", args),
			zap.Duration("duration", time.Since(now)),
		)
	}
}

// TODO: bring logger hook as in rdb package https://github.com/yanakipre/bot/issues/7382
func (db *DB) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	defer makeLoggerHook(ctx, query, args)()
	return db.conn.Select(ctx, dest, query, args...)
}

func (db *DB) PrepareBatchContext(ctx context.Context, query string) (driver.Batch, error) {
	return db.conn.PrepareBatch(ctx, query)
}

func (db *DB) InsertContext(ctx context.Context, query string, args ...any) error {
	defer makeLoggerHook(ctx, query, args)()
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
	defer makeLoggerHook(ctx, query, args)()
	return db.conn.Exec(ctx, query, args...)
}

// ExecMultiContext executes query with multiple statements.
// clickhouse-go client doesn't support it automatically,
// so do as golang-migrate - split query by `;` and execute each statement.
func (db *DB) ExecMultiContext(ctx context.Context, query string) error {
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
