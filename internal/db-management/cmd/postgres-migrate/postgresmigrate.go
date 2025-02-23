package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/pgtooling"
	"go.uber.org/zap"
)

func main() {
	flag.Parse()
	ctx, defaultBehaviourForSignals := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer defaultBehaviourForSignals()
	logger.SetNewGlobalLoggerOnce(logger.DefaultConfig())
	lg := logger.FromContext(ctx)
	defer lg.Info("application shutting down")

	config, err := pgx.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Fatal(ctx, "cannot create pgx config", zap.Error(err))
	}

	opts := pgtooling.MigrateOpts{
		Destination:        flagDestination,
		PathToDBDir:        ".",
		SchemaVersionTable: tableName,
	}

	opts.SetDefaults()

	if err := pgtooling.Migrate(ctx, config, opts); err != nil {
		switch {
		case errors.Is(err, pgtooling.ErrDbVersionIsNewer):
			// Suppose the following case:
			// 	1. We have a release that does database migrations. It updates versions table successfully.
			// 	2. Then we need to roll back the release to previous version.
			//	The previous version contains set of migration that is smaller than the current version.
			//
			// That is when we get this error.
			//
			// But that is OK, we don't need to error out: our migrations are always backwards compatible,
			// so old version of code is supposed to work.
			lg.Warn(
				"database version is newer than migrations that we have, that is OK only during rollback",
				zap.Error(err),
			)
		default:
			lg.Fatal("cannot migrate", zap.Error(err))
		}
		return
	}
	lg.Info("successfully migrated")
}

var (
	lockTimeout     string
	flagDestination string
	migrationsDir   string
	tableName       string
)

func init() {
	flag.StringVar(
		&lockTimeout,
		"locktimeout",
		"5s",
		"value for SET lock_timeout. Consider lower values to do not block others in the lock queue for long time.",
	)
	flag.StringVar(&flagDestination, "destination", "last", "migration destination")
	flag.StringVar(&tableName, "table", "public.schema_version", "table name with migration version")
}
