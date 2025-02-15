package rdb

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kamilsk/retry/v5"
	"github.com/kamilsk/retry/v5/strategy"
	"github.com/orlangure/gnomock"
	"github.com/rekby/fixenv"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/rdb/internal/driver"
	"github.com/yanakipre/bot/internal/rdb/internal/pgretries"
	"github.com/yanakipre/bot/internal/rdb/rdbtesttooling"
)

// TestMain is required to setup rdb test package
func TestMain(m *testing.M) {
	var exitCode int

	cfg := logger.DefaultConfig()
	cfg.Format = logger.FormatConsole
	cfg.LogLevel = "DEBUG"
	logger.SetNewGlobalLoggerQuietly(cfg)

	// initialize package env
	_, cancel := fixenv.CreateMainTestEnv(nil)
	defer func() {
		cancel()
		os.Exit(exitCode)
	}()

	exitCode = m.Run()
}

var testErr = errors.New("test error")

func TestDB_WithTx(t *testing.T) {
	if testing.Short() {
		t.Skip("skip long-running test in short mode")
	}
	t.Parallel()

	type args struct {
		ctx context.Context
		o   TxOptions
	}
	tests := []struct {
		name    string
		args    func(ctx context.Context) args
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "some user defined error is propagated back to user",
			args: func(ctx context.Context) args {
				return args{
					ctx: ctx,
					o: TxOptions{
						ContinueActive: false,
						F: func(ctx context.Context, tx *sqlx.Tx) error {
							return fmt.Errorf("this should be propagated to the user: %w", testErr)
						},
					},
				}
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorIs(t, err, testErr)
			},
		},
		{
			name: "context cancellation on TX rollback results in context.Canceled",
			args: func(ctx context.Context) args {
				ctx, cancel := context.WithTimeout(ctx, time.Second*10)
				return args{
					ctx: ctx,
					o: TxOptions{
						ContinueActive: false,
						F: func(ctx context.Context, tx *sqlx.Tx) error {
							cancel() // this prevents rollback from ending successfully
							return errors.New("this should immediately result in rollback")
						},
					},
				}
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorIs(t, err, context.Canceled)
			},
		},
		{
			name: "context cancellation on TX commit results in context canceled",
			args: func(ctx context.Context) args {
				ctx, cancel := context.WithTimeout(ctx, time.Second*1)
				return args{
					ctx: ctx,
					o: TxOptions{
						ContinueActive: false,
						F: func(ctx context.Context, tx *sqlx.Tx) error {
							cancel()
							select {
							case <-ctx.Done():
								logger.Info(ctx, "context is done, returning nil")
								return nil // we should get there
							case <-time.After(time.Second * 10):
								logger.Info(ctx, "hit timeout, returning error")
								return errors.New("could not get the context cancellation")
							}
						},
					},
				}
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorIs(t, err, context.Canceled)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connstr := rdbtesttooling.FixturePostgresProject(fixenv.New(t))
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			cfg := DefaultConfig()
			cfg.DSN = connstr
			db := New(cfg)
			require.NoError(t, db.Ready(ctx))
			a := tt.args(ctx)
			err := db.WithTx(a.ctx, a.o)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			logger.Error(ctx, fmt.Errorf("got error: %w", err))
			tt.wantErr(t, err)
		})
	}
}

func TestOpen(t *testing.T) {
	if testing.Short() {
		t.Skip("skip long-running test in short mode")
	}
	ctx := context.Background()
	logger.SetNewGlobalLoggerQuietly(logger.DefaultConfig())
	t.Run("by default when no server - error", func(t *testing.T) {
		// should not retry and propagate the error
		ctx := logger.WithName(ctx, t.Name())
		connCfg, err := pgx.ParseConfig(fmt.Sprintf("postgres://postgres:password@localhost:%d/postgres", getFreePort()))
		require.NoError(t, err)

		opts := []driver.OpenOption{
			driver.WithLogging(zapcore.DebugLevel),
		}
		dr, err := driver.Open(
			driverName,
			stdlib.RegisterConnConfig(connCfg),
			opts...,
		)

		session := &sqlxdbWrapper{DB: sqlx.NewDb(dr, driverName), retries: pgretries.NoOp()}
		db := New(DefaultConfig())
		db.session = session
		// don't check for readiness - no server is running
		require.NoError(t, err)
		_, err = db.Exec(ctx, "test", "SELECT 1", map[string]any{})
		require.Error(t, err)
		e := &pgconn.ConnectError{}
		require.True(t, errors.As(err, &e), "got %+v instead of expected error", err)
	})
	t.Run("no server, pgretries stragegy and 2 retries", func(t *testing.T) {
		// should stop retrying on third attempt
		ctx := logger.WithName(ctx, t.Name())

		tried := 0

		how := pgretries.Strategy()
		how = append(how, func(_ retry.Breaker, attempt uint, err error) bool {
			// count how many times the strategies declared before this one, allowed to execute retries
			// including the very first attempt
			tried += 1
			logger.Info(ctx, "attempt might happen depending on following strategies", zap.Uint("attempt", attempt))
			return true
		})
		how = append(how, strategy.Limit(2))

		opts := []driver.OpenOption{
			driver.WithLogging(zapcore.DebugLevel),
		}

		connCfg, err := pgx.ParseConfig(fmt.Sprintf("postgres://postgres:password@localhost:%d/postgres", getFreePort()))
		require.NoError(t, err)

		dr, err := driver.Open(
			driverName,
			stdlib.RegisterConnConfig(connCfg),
			opts...,
		)
		require.NoError(t, err)
		session := &sqlxdbWrapper{DB: sqlx.NewDb(dr, driverName), retries: how}
		db := &DB{session: session, tracer: otel.Tracer("rdb")}
		_, err = db.Exec(ctx, "test", "SELECT 1", map[string]any{})
		require.Error(t, err)
		require.Equal(t, 3, tried)
		e := &pgconn.ConnectError{}
		require.True(t, errors.As(err, &e), "got %+v instead of expected error", err)
	})
	t.Run("server exists and pgretries strategy", func(t *testing.T) {
		// should retry tx
		c, err := rdbtesttooling.StartReusablePGContainer("14-alpine", "reusable-test-db")

		require.NoError(t, err)
		ctx := logger.WithName(ctx, t.Name())
		tried := 0

		how := pgretries.Strategy()
		how = append(how, func(_ retry.Breaker, attempt uint, err error) bool {
			// count how many times the strategies declared before this one, allowed to execute retries
			// including the very first attempt
			tried += 1
			logger.Info(ctx, "attempt might happen depending on following strategies", zap.Uint("attempt", attempt))
			return true
		})
		how = append(how, strategy.Limit(2))

		opts := []driver.OpenOption{
			driver.WithLogging(zapcore.DebugLevel),
		}

		connCfg, err := pgx.ParseConfig(fmt.Sprintf("postgres://postgres:password@%s:%d/postgres", c.Host, c.DefaultPort()))
		require.NoError(t, err)

		dr, err := driver.Open(
			driverName,
			stdlib.RegisterConnConfig(connCfg),
			opts...,
		)
		require.NoError(t, err)
		session := &sqlxdbWrapper{DB: sqlx.NewDb(dr, driverName), retries: how}
		db := &DB{session: session, tracer: otel.Tracer("rdb")}
		_, err = db.Exec(ctx, "test", "SELECT 1", map[string]any{})
		require.NoError(t, err)
		require.Equal(t, 1, tried)
	})
	t.Run("no server exists and tx retries", func(t *testing.T) {
		// should connect from the first attempt
		ctx := logger.WithName(ctx, t.Name())
		tried := 0

		how := pgretries.Strategy()
		how = append(how, func(_ retry.Breaker, attempt uint, err error) bool {
			// count how many times the strategies declared before this one, allowed to execute retries
			// including the very first attempt
			tried += 1
			logger.Info(ctx, "attempt might happen depending on following strategies", zap.Uint("attempt", attempt))
			return true
		})
		how = append(how, strategy.Limit(2))

		opts := []driver.OpenOption{
			driver.WithLogging(zapcore.DebugLevel),
		}

		connCfg, err := pgx.ParseConfig(fmt.Sprintf("postgres://postgres:password@localhost:%d/postgres", getFreePort()))
		require.NoError(t, err)

		dr, err := driver.Open(
			driverName,
			stdlib.RegisterConnConfig(connCfg),
			opts...,
		)
		require.NoError(t, err)
		session := &sqlxdbWrapper{DB: sqlx.NewDb(dr, driverName), retries: how}
		db := New(DefaultConfig()) //  &DB{session: session, tracer: otel.Tracer("rdb")}
		db.session = session
		db.retries = how
		err = db.WithTx(ctx, TxOptions{
			FIsIdempotent: true,
			F: func(ctx context.Context, tx *sqlx.Tx) error {
				panic("not supposed to get here")
			},
		})
		require.Error(t, err)
		require.Equal(t, 3, tried)
	})
	t.Run("server exists, no need to retry tx", func(t *testing.T) {
		// should connect from the first attempt
		ctx := logger.WithName(ctx, t.Name())
		c, err := rdbtesttooling.StartReusablePGContainer("14-alpine", "reusable-test-db")

		tried := 0

		how := pgretries.Strategy()
		how = append(how, func(_ retry.Breaker, attempt uint, err error) bool {
			// count how many times the strategies declared before this one, allowed to execute retries
			// including the very first attempt
			tried += 1
			logger.Info(ctx, "attempt might happen depending on following strategies", zap.Uint("attempt", attempt))
			return true
		})
		how = append(how, strategy.Limit(2))

		opts := []driver.OpenOption{
			driver.WithLogging(zapcore.DebugLevel),
		}

		connCfg, err := pgx.ParseConfig(fmt.Sprintf("postgres://postgres:password@%s:%d/postgres", c.Host, c.DefaultPort()))

		require.NoError(t, err)

		dr, err := driver.Open(
			driverName,
			stdlib.RegisterConnConfig(connCfg),
			opts...,
		)
		require.NoError(t, err)
		session := &sqlxdbWrapper{DB: sqlx.NewDb(dr, driverName), retries: how}
		db := New(DefaultConfig()) //  &DB{session: session, tracer: otel.Tracer("rdb")}
		db.session = session
		db.retries = how
		err = db.WithTx(ctx, TxOptions{
			FIsIdempotent: true,
			F: func(ctx context.Context, tx *sqlx.Tx) error {
				_, err := db.Exec(ctx, "test", "SELECT 1", map[string]any{})
				return err
			},
		})
		require.NoError(t, err)
		require.Equal(t, 1, tried)
	})
	t.Run("existing server shutdown when tx running", func(t *testing.T) {
		// should connect from the first attempt
		ctx := logger.WithName(ctx, t.Name())
		c, err := rdbtesttooling.StartReusablePGContainer("14-alpine", "reusable-test-db")

		tried := 0

		how := pgretries.Strategy()
		how = append(how, func(_ retry.Breaker, attempt uint, err error) bool {
			// count how many times the strategies declared before this one, allowed to execute retries
			// including the very first attempt
			tried += 1
			logger.Info(ctx, "attempt might happen depending on following strategies", zap.Uint("attempt", attempt))
			return true
		})
		how = append(how, strategy.Limit(2))

		opts := []driver.OpenOption{
			driver.WithLogging(zapcore.DebugLevel),
		}

		connCfg, err := pgx.ParseConfig(fmt.Sprintf("postgres://postgres:password@%s:%d/postgres", c.Host, c.DefaultPort()))

		require.NoError(t, err)

		dr, err := driver.Open(
			driverName,
			stdlib.RegisterConnConfig(connCfg),
			opts...,
		)
		require.NoError(t, err)
		session := &sqlxdbWrapper{DB: sqlx.NewDb(dr, driverName), retries: how}
		db := New(DefaultConfig()) //  &DB{session: session, tracer: otel.Tracer("rdb")}
		db.session = session
		db.retries = how
		select1Succeeded, select2Succeeded := 0, 0
		err = db.WithTx(ctx, TxOptions{
			FIsIdempotent: true,
			F: func(ctx context.Context, tx *sqlx.Tx) error {
				// SELECT 1 call will pass
				if _, err := db.Exec(ctx, "test-select-1", "SELECT 1", map[string]any{}); err != nil {
					panic(err) // not supposed to happen
				}
				select1Succeeded += 1

				if err := gnomock.Stop(c); err != nil {
					panic(err) // not supposed to happen
				}

				// SELECT 2 call will NOT pass and there will be a retry attempt
				_, err := db.Exec(ctx, "test-select-2", "SELECT 2", map[string]any{})
				if err == nil {
					select2Succeeded += 1
				}
				return err
			},
		})
		require.Error(t, err)
		// this should be strictly 3, always, because we set strategy.Limit(2),
		// and counter is executed before strategy.Limit(2).
		// > 3 means two independent retries took place.
		require.Equal(t, 3, tried)
		require.Equal(t, select1Succeeded, 1)
		require.Equal(t, select2Succeeded, 0)
	})
}

func getFreePort() int {
	// Listen on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Errorf("failed to find a free port: %w", err))
	}
	defer listener.Close()

	// Extract the port number from the listener
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port
}
