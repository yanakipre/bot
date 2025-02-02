package rdb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rekby/fixenv"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
	"github.com/yanakipe/bot/internal/rdb/rdbtesttooling"
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
		ctx            context.Context
		continueActive bool
		f              func(ctx context.Context, tx *sqlx.Tx) error
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
					ctx:            ctx,
					continueActive: false,
					f: func(ctx context.Context, tx *sqlx.Tx) error {
						return fmt.Errorf("this should be propagated to the user: %w", testErr)
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
					ctx:            ctx,
					continueActive: false,
					f: func(ctx context.Context, tx *sqlx.Tx) error {
						cancel() // this prevents rollback from ending successfully
						return errors.New("this should immediately result in rollback")
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
					ctx:            ctx,
					continueActive: false,
					f: func(ctx context.Context, tx *sqlx.Tx) error {
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
				}
			},
			wantErr: func(t require.TestingT, err error, i ...any) {
				require.ErrorIs(t, err, context.Canceled)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connstr := rdbtesttooling.FixturePostgresProject(fixenv.NewEnv(t))
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			cfg := DefaultConfig()
			cfg.DSN = connstr
			db := New(cfg)
			require.NoError(t, db.Ready(ctx))
			a := tt.args(ctx)
			err := db.WithTx(a.ctx, a.continueActive, a.f)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			logger.Error(ctx, "got error", zap.Error(err))
			tt.wantErr(t, err)
		})
	}
}
