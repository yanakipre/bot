package pgtooling

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kamilsk/retry/v5"
	"github.com/kamilsk/retry/v5/strategy"
	"github.com/orlangure/gnomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/projectpath"
	"github.com/yanakipre/bot/internal/rdb/rdbtesttooling"
	"github.com/yanakipre/bot/internal/secret"
	"github.com/yanakipre/bot/internal/testtooling"
)

func TestFormattedPgDump(t *testing.T) {
	testtooling.SkipShort(t)
	type args struct {
		ddl string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "create table",
			args: args{ddl: "CREATE TABLE a (id INT)"},
			want: "CREATE TABLE public.a (\n    id integer\n);",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			container, err := rdbtesttooling.StartPGContainer(ctx)
			require.NoError(t, err)
			defer gnomock.Stop(container)
			dsn := secret.NewString(fmt.Sprintf(
				"postgres://postgres:password@%s:%d/postgres",
				container.Host,
				container.DefaultPort(),
			))
			connect, err := pgx.Connect(ctx, dsn.Unmask())
			require.NoError(t, err)
			_, err = connect.Exec(ctx, tt.args.ddl)
			require.NoError(t, err)
			got, err := FormattedPgDump(dsn)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

var pathToFixtures = path.Join(projectpath.RootPath, "/internal/pgtooling/internal/fixtures/")

func TestMigrate(t *testing.T) {
	testtooling.SkipShort(t)
	testtooling.SetNewGlobalLoggerQuietly()
	tests := []struct {
		name  string
		tCase func(ctx context.Context, t *testing.T, pgContainer *gnomock.Container, dsn string)
	}{
		{
			name: "db-locks migrates correctly",
			tCase: func(ctx context.Context, t *testing.T, pgContainer *gnomock.Container, dsn string) {
				connect, err := pgx.Connect(ctx, dsn)
				require.NoError(t, err)
				defer connect.Close(ctx)
				dbSQL, err := os.ReadFile(path.Join(pathToFixtures, "db-locks/db.sql"))
				if err != nil {
					return
				}
				_, err = connect.Exec(ctx, string(dbSQL))
				require.NoError(t, err)

				err = Migrate(ctx, connect.Config(), MigrateOpts{
					PathToDBDir: path.Join(pathToFixtures, "db-locks"),
					Destination: "last",
				})
				require.NoError(t, err)

				row := connect.QueryRow(ctx, "SELECT COUNT(*) FROM table_for_locks")
				var count int = -1
				err = row.Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 1, count)
			},
		},
		{
			name: "holding lock forever results in ErrLockNotAcquired",
			tCase: func(ctx context.Context, t *testing.T, pgContainer *gnomock.Container, dsn string) {
				connect, err := pgx.Connect(ctx, dsn)
				require.NoError(t, err)
				defer connect.Close(ctx)
				dbSQL, err := os.ReadFile(path.Join(pathToFixtures, "db-locks/db.sql"))
				if err != nil {
					return
				}
				_, err = connect.Exec(ctx, string(dbSQL))
				require.NoError(t, err)

				// hold a lock in a transaction
				lockConn, err := pgx.Connect(ctx, dsn)
				require.NoError(t, err)
				defer lockConn.Close(ctx)
				tx, err := lockConn.BeginTx(ctx, pgx.TxOptions{})
				require.NoError(t, err)
				_, err = tx.Exec(ctx, "SELECT * FROM table_for_locks")
				require.NoError(t, err)

				err = Migrate(ctx, connect.Config(), MigrateOpts{
					PathToDBDir: path.Join(pathToFixtures, "db-locks"),
					Destination: "last",
					LockTimeout: "50ms",
					Retry: retry.How{
						strategy.Limit(10),
					},
				})
				require.ErrorIs(t, err, ErrLockNotAcquired)
			},
		},
		{
			name: "holding and releasing lock results in success",
			tCase: func(ctx context.Context, t *testing.T, pgContainer *gnomock.Container, dsn string) {
				connect, err := pgx.Connect(ctx, dsn)
				require.NoError(t, err)
				defer connect.Close(ctx)
				dbSQL, err := os.ReadFile(path.Join(pathToFixtures, "db-locks/db.sql"))
				if err != nil {
					return
				}
				_, err = connect.Exec(ctx, string(dbSQL))
				require.NoError(t, err)

				// hold a lock in a transaction but release it after Migrate retries
				lockConn, err := pgx.Connect(ctx, dsn)
				require.NoError(t, err)
				defer lockConn.Close(ctx)
				tx, err := lockConn.BeginTx(ctx, pgx.TxOptions{})
				require.NoError(t, err)
				_, err = tx.Exec(ctx, "SELECT * FROM table_for_locks")
				require.NoError(t, err)

				var unlockStrategy strategy.Strategy = func(breaker retry.Breaker, attempt uint, err error) bool {
					// allows one retry, then releases the lock
					logger.Info(ctx, "attempt", zap.Any("number", attempt), zap.Error(err))
					switch {
					case attempt == 2:
						lockConn.Close(ctx)
					case attempt > 2:
						panic("cannot happen")
					}
					return true
				}

				err = Migrate(ctx, connect.Config(), MigrateOpts{
					PathToDBDir: path.Join(pathToFixtures, "db-locks"),
					Destination: "last",
					LockTimeout: "50ms",
					Retry:       retry.How{unlockStrategy},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "db-anonymous-blocks migrates correctly",
			tCase: func(ctx context.Context, t *testing.T, pgContainer *gnomock.Container, dsn string) {
				connect, err := pgx.Connect(ctx, dsn)
				require.NoError(t, err)
				defer connect.Close(ctx)
				dbSQL, err := os.ReadFile(path.Join(pathToFixtures, "db-anonymous-blocks/db.sql"))
				if err != nil {
					return
				}
				_, err = connect.Exec(ctx, string(dbSQL))
				require.NoError(t, err)

				err = Migrate(ctx, connect.Config(), MigrateOpts{
					PathToDBDir: path.Join(pathToFixtures, "db-anonymous-blocks"),
					Destination: "last",
				})
				require.NoError(t, err)

				row := connect.QueryRow(ctx, `SELECT COUNT(*) FROM branches WHERE "default"`)
				var count int = -1
				err = row.Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 4, count)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			container, err := rdbtesttooling.StartPGContainer(ctx)
			require.NoError(t, err)
			defer gnomock.Stop(container)
			dsn := fmt.Sprintf(
				"postgres://postgres:password@%s:%d/postgres",
				container.Host,
				container.DefaultPort(),
			)
			now := time.Now()
			tt.tCase(ctx, t, container, dsn)
			logger.Info(ctx, "finished", zap.Duration("took", time.Since(now)))
		})
	}
}
