// Package rdb stands for relational database.
// It is a driver to access postgres.
package rdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kamilsk/retry/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/rdb/internal/driver"
	"github.com/yanakipre/bot/internal/rdb/internal/driver/sqldata"
	"github.com/yanakipre/bot/internal/rdb/internal/pgretries"
	"github.com/yanakipre/bot/internal/secret"
)

type DB struct {
	session   *sqlxdbWrapper
	conf      Config
	tracer    trace.Tracer
	retries   retry.How
	noRetries retry.How
}

func (d *DB) Ready(ctx context.Context) error {
	lg := logger.FromContext(ctx)
	if d.session != nil {
		lg.Info("db.Ready called more than once, not an error, but not expected")
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
	d.session = &sqlxdbWrapper{DB: session, retries: d.retries}
	return nil
}

// ExposeDriverFromDB is highly not recommended for usage.
// This is a temporary function to work on River Queue because it needs access to the sql.DB.
// TODO: #21704 remove
// TODO: #21704 ensure job is scheduled and finished in a transactional way
//
// DEPRECATED
func ExposeDriverFromDB(db *DB) *sql.DB {
	return db.session.DB.DB
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

type TxOptions struct {
	// ContinueActive set to true will reuse existing tx or start new one.
	// Setting it to false forbids nested transactions, resulting in an error when tx is in flight.
	ContinueActive bool
	// FIsIdempotent set to true will rerun the whole transaction in case of retryable error.
	// User must supply an idempotent F function, that does not depend on outer function scope.
	// See DB.WithTx for more details.
	FIsIdempotent bool
	F             func(ctx context.Context, tx *sqlx.Tx) error
}

// WithTx starts transaction or continues the existing one.
// Only beginTx statement is retried. We could retry the whole transaction statement,
// BUT WithTx accepts f, which can see the outer scope and modify it.
// In general there is no guarantee the user writes an f function that is idempotent.
// Imagine a case:
//
// var someVar int
//
//	err := db.WithTx(ctx, false, func(ctx context.Context, tx *sqlx.Tx) error {
//		if someVar == 0 {
//	      someVar++
//		}
//	    ...some query that will catch a retryable error...
//	    return nil
//	})
//
//	if someVar != 1 {
//		panic("someVar should be 1")
//
// This illustrates the problem. If we retry the whole transaction, someVar will be 2.
// So we retry only the beginTx statement unless user explicitly sets FIsIdempotent to true.
func (d *DB) WithTx(ctx context.Context, o TxOptions) error {
	activeTx, ok := ctx.Value(TxKey).(*sqlx.Tx)
	if ok && !o.ContinueActive {
		return errors.New("failed to begin new tx, tx already exists")
	}

	// we're inside active transaction, continue
	if ok && o.ContinueActive {
		return o.F(ctx, activeTx)
	}

	var p []TxCallback
	ctx = context.WithValue(ctx, afterCommitCallbacksKey, &p)

	// no active transaction
	retries := d.noRetries
	if o.FIsIdempotent {
		// The whole transaction is idempotent, we can retry it
		retries = d.retries
	}
	err := retry.Do(ctx, func(ctx context.Context) error {
		tx, finishTx, err := d.beginTx(ctx)
		if err != nil {
			return err
		}
		ctx = context.WithValue(ctx, TxRollbackKey, tx.Rollback)

		// insert tx into context and pass it to callback
		ctx, cancel := context.WithCancel(context.WithValue(ctx, TxKey, tx))
		defer cancel()

		err = finishTx(o.F(ctx, tx))
		return err
	}, retries...)

	return err
}

var errTransactionAborted = errors.New("transaction was aborted")

func (d *DB) beginTx(ctx context.Context) (*sqlx.Tx, func(error) error, error) {
	tx, err := d.session.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("could not open transaction: %w", err)
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
			// context is not canceled yet, but it might happen during rollback
			e := tx.Rollback()
			if e != nil {
				if !errors.Is(e, sql.ErrTxDone) {
					err = fmt.Errorf("failed to rollback transaction after error: %w: %w", err, e)
				} else if ctx.Err() != nil {
					// context cancellation happened during rollback,
					// we return it explicitly
					err = fmt.Errorf("failed to rollback transaction after error: %w: %w", err, ctx.Err())
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
			logger.Error(ctx, errors.New("afterCommitCallbacks array not found in context"))
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

func (d *DB) Get(
	ctx context.Context,
	dest any,
	name string,
	query string,
	arg any,
) error {
	ctx = sqldata.NewContext(ctx, sqldata.Operation(name))

	ctx, span := d.tracer.Start(ctx,
		"Postgres Get",
		trace.WithAttributes(
			attribute.String("query", query),
		),
	)
	defer span.End()

	tx := d.TxExecutorFromCtx(ctx)
	q, args, err := tx.BindNamed(query, arg)
	if err != nil {
		return err
	}

	// validate dest is non-nil pointer
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New(fmt.Sprint("dest arg for ", name, "must be a non-nil pointer"))
	}

	// if we expect a slice use "select", otherwise use "get"
	v = reflect.Indirect(v)

	// some slice types implement scanner, use "get" on them
	t := reflect.TypeOf(v)
	s := reflect.TypeOf((*sql.Scanner)(nil)).Elem()

	if v.Kind() == reflect.Slice && !t.Implements(s) {
		return tx.SelectContext(ctx, dest, q, args...)
	} else {
		return tx.GetContext(ctx, dest, q, args...)
	}
}

func (d *DB) Exec(
	ctx context.Context,
	name string,
	query string,
	arg any,
) (sql.Result, error) {
	ctx = sqldata.NewContext(ctx, sqldata.Operation(name))

	ctx, span := d.tracer.Start(ctx,
		"Postgres Exec",
		trace.WithAttributes(
			attribute.String("query", query),
		),
	)
	defer span.End()

	tx := d.TxExecutorFromCtx(ctx)
	q, args, err := tx.BindNamed(query, arg)
	if err != nil {
		return nil, err
	}

	return tx.ExecContext(ctx, q, args...)
}

func New(cfg Config) *DB {
	tracer := otel.Tracer("rdb")
	return &DB{conf: cfg, tracer: tracer, retries: pgretries.Strategy(), noRetries: pgretries.NoOp()}
}
