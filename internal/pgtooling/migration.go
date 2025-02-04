package pgtooling

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
	"github.com/kamilsk/retry/v5"
	"github.com/kamilsk/retry/v5/backoff"
	"github.com/kamilsk/retry/v5/strategy"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/secret"
)

var (
	ErrLockNotAcquired  = errors.New("lock not acquired within given time")
	ErrDbVersionIsNewer = errors.New("database version is newer than image database version")
)

// Migrate allows to migrate the database given the database directory.
func Migrate(
	ctx context.Context,
	connCfg *pgx.ConnConfig,
	opts MigrateOpts,
) error {
	lg := logger.FromContext(ctx)
	opts.SetDefaults()
	// within the retry func,
	// set this to the needed error error and return nil
	// to return the needed error
	var returnErr error
	err := retry.Do(ctx, func(ctx context.Context) error {
		err := migrateWithTern(
			ctx,
			connCfg,
			opts,
		)
		if err != nil {
			pgErr := migrate.MigrationPgError{}
			bvErr := migrate.BadVersionError("")
			switch {
			case errors.As(err, &pgErr):
				if pgErr.Code == "55P03" {
					// 55P03 	lock_not_available
					// https://www.postgresql.org/docs/current/errcodes-appendix.html
					lg.Warn(
						"lock not acquired",
						zap.String("migration_name", pgErr.MigrationName),
						zap.String("stmt", pgErr.Sql),
					)
					return errors.Join(ErrLockNotAcquired, bvErr) // retry error when lock cannot be taken
				}
			case errors.As(err, &bvErr):
				if strings.Contains(bvErr.Error(), "current version") {
					returnErr = errors.Join(ErrDbVersionIsNewer, bvErr)
					return nil // this should not be retried
				}
			}
			returnErr = err
			return nil // this should not be retried
		}
		return nil
	}, opts.Retry...)
	if err != nil {
		return err
	}
	return returnErr
}

func pgDump(args ...string) (string, error) {
	cmdArgs := []string{"run", "--network", "host", "postgres:14-alpine", "pg_dump"}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command("docker", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pg_dump failed with: %v\noutput:\n%v\n%w", err, string(output), err)
	}
	return string(output), nil
}

var omitLineStart = []string{
	"SELECT ",
	"SET ",
	"--",
}

func omitByLineStart(line string) bool {
	for _, prefix := range omitLineStart {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

func omitBySubstring(line string) bool {
	return strings.Contains(line, "OWNER TO postgres")
}

func filterOutPGDumpInfo(info string) string {
	var result []string
	omittingEmpty := false
	for _, line := range strings.Split(info, "\n") {
		if omitByLineStart(line) {
			continue
		}
		if omitBySubstring(line) {
			continue
		}
		if line == "" {
			if omittingEmpty {
				continue
			}
			omittingEmpty = true
		} else if omittingEmpty {
			omittingEmpty = false
		}
		result = append(result, line)
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}

func FormattedPgDump(dsn secret.String) (string, error) {
	dump, err := pgDump(
		"--dbname",
		dsn.Unmask(),
		"--schema-only",
		"--no-comments",
		"--no-tablespaces",
	)
	if err != nil {
		if strings.Contains(
			err.Error(),
			"Is the server running on that host and accepting TCP/IP connections?",
		) {
			// try "host.docker.internal"
			if attemptedDump, attemptedErr := pgDump(
				"--dbname",
				strings.Replace(dsn.Unmask(), "127.0.0.1", "host.docker.internal", 1),
				"--schema-only",
				"--no-comments",
				"--no-tablespaces",
			); attemptedErr != nil {
				return "", errors.Join(attemptedErr, err)
			} else {
				// success
				dump = attemptedDump
			}
		} else {
			return "", err
		}
	}
	return filterOutPGDumpInfo(dump), nil
}

type MigrateOpts struct {
	PathToDBDir string
	// Migration Destination in format acceptable by tern library.
	// Usually it's "last" or a positive number.
	Destination string
	// LockTimeout is passed as is into SET lock_timeout statement.
	// Should be in format postgres accepts.
	// Example: 1s
	LockTimeout string
	// Retry strategy for cases when it's possible.
	// Especially, when lock cannot be acquired after LockTimeout.
	Retry retry.How
}

func (m *MigrateOpts) SetDefaults() {
	if m.Destination == "" {
		m.Destination = "last"
	}
	if m.LockTimeout == "" {
		m.LockTimeout = "5s"
	}
	if m.Retry == nil {
		m.Retry = retry.How{
			// By default, wait 1 second, and try again with incremental backoff of 1 second.
			strategy.Backoff(backoff.Incremental(time.Millisecond*500, time.Millisecond*500)),
			strategy.Limit(10),
		}
	}
}

// migrateWithTern is a copy of https://github.com/jackc/tern/blob/master/main.go
func migrateWithTern(
	ctx context.Context,
	connCfg *pgx.ConnConfig,
	opts MigrateOpts,
) error {
	conn, err := pgx.ConnectConfig(ctx, connCfg)
	if err != nil {
		return fmt.Errorf("could not connect to postgres database: %w", err)
	}

	_, err = conn.Exec(ctx, fmt.Sprintf("SET lock_timeout='%s'", opts.LockTimeout))
	if err != nil {
		return fmt.Errorf("could not set lock_timeout: %w", err)
	}

	defer func() {
		if err := conn.Close(ctx); err != nil {
			logger.Warn(ctx, "could not close database connection", zap.Error(err))
		}
	}()

	migrator, err := migrate.NewMigrator(ctx, conn, "public.schema_version")
	if err != nil {
		return fmt.Errorf("could not create migrator: %w", err)
	}
	pathToMigrations := strings.TrimRight(opts.PathToDBDir, "/") + "/migrations/"
	err = migrator.LoadMigrations(os.DirFS(pathToMigrations))
	if err != nil {
		return fmt.Errorf("could not load migrations in %q: %w", pathToMigrations, err)
	}
	if len(migrator.Migrations) == 0 {
		return errors.New("no migrations found")
	}
	migrator.OnStart = func(sequence int32, name, direction, sql string) {
		logger.Debug(ctx, "executing", zap.String("migration_name", name), zap.String("direction", direction))
	}
	var currentVersion int32
	currentVersion, err = migrator.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("could not get current version: %w", err)
	}

	mustParseDestination := func(d string) int32 {
		var n int64
		n, err = strconv.ParseInt(d, 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bad destination:\n  %v\n", err)
			os.Exit(1)
		}
		return int32(n)
	}

	destination := opts.Destination
	if destination == "last" {
		err = migrator.Migrate(ctx)
	} else if len(destination) >= 3 && destination[0:2] == "-+" {
		err = migrator.MigrateTo(ctx, currentVersion-mustParseDestination(destination[2:]))
		if err == nil {
			err = migrator.MigrateTo(ctx, currentVersion)
		}
	} else if len(destination) >= 2 && destination[0] == '-' {
		err = migrator.MigrateTo(ctx, currentVersion-mustParseDestination(destination[1:]))
	} else if len(destination) >= 2 && destination[0] == '+' {
		err = migrator.MigrateTo(ctx, currentVersion+mustParseDestination(destination[1:]))
	} else {
		err = migrator.MigrateTo(ctx, mustParseDestination(destination))
	}
	return err
}
