// Package rdb stands for relational database.
// It is a driver to access postgres.
package rdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"

	"github.com/yanakipe/bot/internal/logger"
	"github.com/yanakipe/bot/internal/rdb/internal/driver"
	"github.com/yanakipe/bot/internal/rdb/internal/driver/sqldata"
	"github.com/yanakipe/bot/internal/secret"
)

type DB struct {
	session *sqlx.DB
	conf    Config
	tracer  trace.Tracer
}

func (d *DB) Ready(ctx context.Context) error {
	if d.session != nil {
		logger.Info(ctx, "db.Ready called more than once, not an error, but not expected")
		return nil
	}
	ctx, span := d.tracer.Start(ctx, "Postgres connect")
	defer span.End()

	if err := d.conf.CheckAndSetDefaults(); err != nil {
		return err
	}

	connCfg, err := pgx.ParseConfig(d.conf.DSN.Unmask())
	if err != nil {
		return fmt.Errorf("dsn parse error: %w", err)
	}
	// Note: in case of using pgbouncer or pgpool
	//connCfg.PreferSimpleProtocol = true
	//connCfg.BuildStatementCache = func(conn *pgconn.PgConn) stmtcache.Cache {
	//	return stmtcache.New(conn, stmtcache.ModeDescribe, 512)
	//}
	opts := []driver.OpenOption{
		driver.WithLogging(zapcore.DebugLevel),
		driver.WithDSNParsedInfo(d.conf.DSN.Unmask()),
	}
	if d.conf.SearchPath != "" {
		opts = append(opts, driver.WithSearchPath(d.conf.SearchPath))
	}

	db, err := driver.Open(
		driverName,
		stdlib.RegisterConnConfig(connCfg),
		opts...,
	)
	if err != nil {
		return fmt.Errorf("driver open error: %w", err)
	}

	db.SetMaxOpenConns(d.conf.MaxOpenConns)
	db.SetConnMaxLifetime(d.conf.MaxConnLifetime.Duration)

	// Send database metrics
	go func() {
		timer := time.NewTicker(d.conf.CollectMetricsInterval.Duration)
		defer timer.Stop()
		select {
		case <-timer.C:
			sendMetrics(db.Stats(), string(d.conf.DatabaseType))
		case <-ctx.Done():
			return
		}
	}()

	session := sqlx.NewDb(db, driverName)
	if err := session.PingContext(ctx); err != nil {
		return err
	}
	d.session = session
	return nil
}

func (d *DB) StartSpan(
	ctx context.Context,
	spanName string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	return d.tracer.Start(ctx, spanName, opts...)
}

// DSN returns DSN from the config.
func (d *DB) DSN() secret.String {
	return d.conf.DSN
}

// TxExecutorFromCtx returns database interface based on provided context. If context
// has running transaction, it'll be returned and queries will be executed inside.
// Otherwise, default sqlx.DB will be used.
func (d *DB) TxExecutorFromCtx(ctx context.Context) TxExecutor {
	currentTx, ok := ctx.Value(TxKey).(*sqlx.Tx)
	if ok {
		return currentTx
	}
	return d.session
}

func (d *DB) MapperFunc(mf func(string) string) {
	d.session.MapperFunc(mf)
}

func (d *DB) Close() error {
	return d.session.Close()
}

func (d *DB) WithTx(
	ctx context.Context,
	continueActive bool,
	f func(ctx context.Context, tx *sqlx.Tx) error,
) error {
	activeTx, ok := ctx.Value(TxKey).(*sqlx.Tx)
	if ok && !continueActive {
		return errors.New("failed to begin new tx, tx already exists")
	}

	// we're inside active transaction, continue
	if ok && continueActive {
		return f(ctx, activeTx)
	}

	var p []TxCallback
	ctx = context.WithValue(ctx, afterCommitCallbacksKey, &p)

	// no active transaction
	tx, finishTx, err := d.beginTx(ctx)
	if err != nil {
		return err
	}

	// insert tx into context and pass it to callback
	ctx, cancel := context.WithCancel(context.WithValue(ctx, TxKey, tx))
	defer cancel()

	err = finishTx(f(ctx, tx))

	return err
}

var errTransactionAborted = errors.New("transaction was aborted")

func (d *DB) beginTx(ctx context.Context) (*sqlx.Tx, func(error) error, error) {
	tx, err := d.session.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, nil, err
	}

	// it commits or rollbacks initialized transaction
	finishTx := func(err error) error {
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				// db connection is already closed by the driver.
				// no need to rollback manually.
				// Instead, we need to explicitly pass the error to the top-level error handlers.
				// see https://pkg.go.dev/database/sql/driver#ConnBeginTx for more details.
				return err
			}
			// context is not cancelled yet, but it might happen during rollback
			e := tx.Rollback()
			if e != nil {
				if !errors.Is(e, sql.ErrTxDone) {
					err = fmt.Errorf("failed to rollback transaction after error %q: %w", err, e)
				} else if ctx.Err() != nil {
					// context cancellation happened during rollback,
					// we return it explicitly
					err = fmt.Errorf("failed to rollback transaction after error %q: %w", err, ctx.Err())
				}
			}
			return err
		}

		e := tx.Commit()
		if errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded) {
			return fmt.Errorf("failed to commit transaction: %w", e)
		} else if errors.Is(e, sql.ErrTxDone) {
			if ctx.Err() != nil {
				// pass the cancellation explicitly first
				// the underlying database/sql even with explicit cancellation works in a not a stable way
				// sometimes the propagated error is sql.ErrTxDone, sometime context cancellation errors.
				// That's the Commit method:
				//
				// database/sql/sql.go

				//	case <-tx.ctx.Done():
				//		if tx.done.Load() {
				//			return ErrTxDone
				//		}
				//		return tx.ctx.Err()
				//	}
				//
				// And there is no priority for awaitDone:
				//
				// database/sql/sql.go
				//
				//  func (tx *Tx) awaitDone() {
				//	  <-tx.ctx.Done()
				//
				// so you can get any of the two errors.
				return fmt.Errorf("failed to commit transaction: %w", ctx.Err())
			}
			return fmt.Errorf("failed to commit rollbacked transaction (timeout): %w", errors.Join(e, errTransactionAborted))
		} else if e != nil {
			return fmt.Errorf("failed to commit transaction: %w", errors.Join(e, errTransactionAborted))
		}

		// Call callbacks
		callbacks, ok := ctx.Value(afterCommitCallbacksKey).(*[]TxCallback)
		if ok {
			for _, callback := range *callbacks {
				callback(ctx)
			}
		} else {
			logger.Error(ctx, "afterCommitCallbacks array not found in context")
		}

		return nil
	}

	return tx, finishTx, err
}

// Register a callback function that will be called after successful commit.
func (d *DB) RegisterAfterCommitCallback(ctx context.Context, f TxCallback) {
	callbacks, ok := ctx.Value(afterCommitCallbacksKey).(*[]TxCallback)
	if ok {
		*callbacks = append(*callbacks, f)
	} else {
		panic("afterCommitCallbacks array not found in context")
	}
}

// Implementation for TxExecutor interface

func (d *DB) DriverName() string {
	return d.session.DriverName()
}

func (d *DB) Rebind(s string) string {
	return d.session.Rebind(s)
}

func (d *DB) BindNamed(s string, i any) (string, []any, error) {
	return d.session.BindNamed(s, i)
}

func (d *DB) QueryContext(
	ctx context.Context,
	query string,
	args ...any,
) (*sql.Rows, error) {
	ctx, span := d.tracer.Start(ctx,
		"Postgres QueryContext",
		trace.WithAttributes(
			attribute.String("query", query),
		),
	)
	defer span.End()
	return d.TxExecutorFromCtx(ctx).QueryContext(ctx, query, args...)
}

func (d *DB) QueryxContext(
	ctx context.Context,
	query string,
	args ...any,
) (*sqlx.Rows, error) {
	ctx, span := d.tracer.Start(ctx,
		"Postgres QueryxContext",
		trace.WithAttributes(
			attribute.String("query", query),
		),
	)
	defer span.End()
	return d.TxExecutorFromCtx(ctx).QueryxContext(ctx, query, args...)
}

func (d *DB) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	ctx, span := d.tracer.Start(ctx,
		"Postgres QueryRowxContext",
		trace.WithAttributes(
			attribute.String("query", query),
		),
	)
	defer span.End()
	return d.TxExecutorFromCtx(ctx).QueryRowxContext(ctx, query, args...)
}

func (d *DB) ExecContext(
	ctx context.Context,
	query string,
	args ...any,
) (sql.Result, error) {
	ctx, span := d.tracer.Start(ctx,
		"Postgres ExecContext",
		trace.WithAttributes(
			attribute.String("query", query),
		),
	)
	defer span.End()
	return d.TxExecutorFromCtx(ctx).ExecContext(ctx, query, args...)
}

func (d *DB) PrepareNamedContext(
	ctx context.Context,
	query string,
	name string,
) (*sqlx.NamedStmt, error) {
	ctx, span := d.tracer.Start(ctx,
		"Postgres PrepareNamedContext",
		trace.WithAttributes(
			attribute.String("query", query),
		),
	)
	defer span.End()
	return d.TxExecutorFromCtx(ctx).
		PrepareNamedContext(sqldata.NewContext(ctx, sqldata.Operation(name)), query)
}

func (d *DB) SelectContext(
	ctx context.Context,
	dest any,
	query string,
	args ...any,
) error {
	ctx, span := d.tracer.Start(ctx,
		"Postgres SelectContext",
		trace.WithAttributes(
			attribute.String("query", query),
		),
	)
	defer span.End()
	return d.TxExecutorFromCtx(ctx).SelectContext(ctx, dest, query, args...)
}

func (d *DB) GetContext(
	ctx context.Context,
	dest any,
	query string,
	args ...any,
) error {
	ctx, span := d.tracer.Start(ctx,
		"Postgres GetContext",
		trace.WithAttributes(
			attribute.String("query", query),
		),
	)
	defer span.End()
	return d.TxExecutorFromCtx(ctx).GetContext(ctx, dest, query, args...)
}

func New(cfg Config) *DB {
	tracer := otel.Tracer("rdb")

	return &DB{conf: cfg, tracer: tracer}
}
