package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // postgres driver
	"github.com/orlangure/gnomock"
	"golang.org/x/crypto/sha3"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/pgtooling"
	"github.com/yanakipre/bot/internal/rdb"
	"github.com/yanakipre/bot/internal/rdb/rdbtesttooling"
	"github.com/yanakipre/bot/internal/secret"
	"github.com/yanakipre/bot/internal/testtooling"
)

func main() {
	logger.SetNewGlobalLoggerOnce(logger.DefaultConfig())

	// This does not work ATM.
	container, err := rdbtesttooling.StartPGContainer("ankane/pgvector@sha256:d3a9d8ac27bb7e05e333ef25b634d2625adaa85336ab729954b9e94859bf6fa7")
	if err != nil {
		panic(fmt.Sprintf("could not start postgres: %v", err))
	}
	defer func() { _ = gnomock.Stop(container) }()
	dsn := secret.NewString(fmt.Sprintf(
		"postgres://postgres:password@%s:%d/postgres",
		container.Host,
		container.DefaultPort(),
	))
	ctx := context.Background()
	cfg := rdb.DefaultConfig()
	cfg.DSN = dsn

	for i, dbPath := range pathsToDBs {
		dbName := dbNames[i]

		logger.Debug(ctx, fmt.Sprintf("migrating db: %s from path: %s", dbName, dbPath))

		teardown, err := migrateSingleDB(ctx, cfg, dbPath, dbName)
		if teardown != nil {
			defer func() { _ = teardown() }()
		}
		if err != nil {
			panic(err)
		}
	}
}

func migrateSingleDB(ctx context.Context, cfg rdb.Config, dbPath, dbName string) (func() error, error) {
	db, teardown, err := testtooling.SetupDBWithName(ctx, cfg, dbName)
	if err != nil {
		return teardown, fmt.Errorf("could not set up db: %w", err)
	}

	connCfg, err := pgx.ParseConfig(db.DSN().Unmask())
	if err != nil {
		return teardown, fmt.Errorf("could not parse config: %w", err)
	}

	opts := pgtooling.MigrateOpts{
		Destination:        "last",
		PathToDBDir:        dbPath,
		SchemaVersionTable: tableName,
	}
	opts.SetDefaults()

	err = pgtooling.Migrate(
		ctx,
		connCfg,
		opts,
	)
	if err != nil {
		return teardown, fmt.Errorf("could not migrate: %w", err)
	}
	currentDbState, err := pgtooling.FormattedPgDump(db.DSN())
	if err != nil {
		return teardown, fmt.Errorf("could not get pg dump: %w", err)
	}
	err = os.WriteFile( //nolint:gosec
		filepath.Join(dbPath, "db.sql"),
		[]byte(currentDbState),
		0o644,
	) //nolint:gosec
	if err != nil {
		return teardown, fmt.Errorf("could not write db.sql: %w", err)
	}

	err = dumpMigrationVersion(dbPath)
	if err != nil {
		return teardown, fmt.Errorf("could not write last_migration.version: %w", err)
	}

	return teardown, nil
}

func dumpMigrationVersion(dbPath string) error {
	// We use git to list the files in the migrations directory
	// to avoid including temporary files, work-in-progress migrations, etc.
	cmd := exec.Command("git", "ls-files", dbPath+"migrations/*.sql")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	// Handle both Windows CRLF and Unix LF line endings
	cleaned := strings.ReplaceAll(string(output), "\r\n", "\n")
	fs := strings.Split(strings.TrimSpace(cleaned), "\n")

	type Migration struct {
		name    string
		version int
	}
	migrations := make([]Migration, 0, len(fs))
	for _, fileName := range fs {
		f := filepath.Base(fileName)
		if f == "events" {
			continue
		}
		version, err := strconv.Atoi(strings.Split(f, "_")[0])
		if err != nil {
			return err
		}
		migrations = append(migrations, Migration{name: f, version: version})
	}
	sort.Slice(
		migrations,
		func(i, j int) bool { return migrations[i].version < migrations[j].version },
	)

	lastMigration := migrations[len(migrations)-1]

	// take hash of file + name to avoid collisions
	bytes, err := os.ReadFile(dbPath + "migrations/" + lastMigration.name) //nolint:gosec
	if err != nil {
		return err
	}
	bytes = append(bytes, []byte(lastMigration.name)...)

	lastMigrationBytes, err := json.Marshal(struct {
		Version int    `json:"version"`
		Hash    string `json:"hash"`
	}{
		Version: lastMigration.version,
		Hash:    fmt.Sprintf("%X", sha3.Sum256(bytes)),
	})
	if err != nil {
		return err
	}

	err = os.WriteFile( //nolint:gosec
		filepath.Join(dbPath, "last_migration.json"),
		lastMigrationBytes,
		0o644,
	) //nolint:gosec
	if err != nil {
		return err
	}

	return nil
}

var (
	pathToDB string

	pathsToDBsStr string
	dbNamesStr    string

	pathsToDBs []string
	dbNames    []string

	tableName string
)

func init() {
	flag.StringVar(&pathToDB, "dbpath", "", "path to database files")
	flag.StringVar(&pathsToDBsStr, "dbpaths", "", "paths to database files")
	flag.StringVar(&dbNamesStr, "dbnames", "", "names of dbs to migrate")
	flag.StringVar(&tableName, "table", "", "table name with migration version")
	flag.Parse()

	if err := validateArgs(); err != nil {
		panic(err)
	}

	if err := parseDBPathsAndNames(); err != nil {
		panic(err)
	}
}

func validateArgs() error {
	if pathToDB != "" && pathsToDBsStr != "" {
		return errors.New("dbpath and dbpaths must not both be set")
	}
	if pathToDB == "" {
		if pathsToDBsStr == "" {
			return errors.New("either dbpath or dbpaths must be set")
		}
		if dbNamesStr == "" {
			return errors.New("dbnames must be set with dbpaths and be of equal length")
		}
	}
	return nil
}

func parseDBPathsAndNames() error {
	if pathsToDBsStr == "" {
		pathsToDBs = []string{pathToDB}
		dbNames = []string{"dbfreeze"}
	} else {
		pathsToDBs = strings.Split(pathsToDBsStr, ",")
		dbNames = strings.Split(dbNamesStr, ",")
	}

	if len(pathsToDBs) != len(dbNames) {
		return errors.New("dbnames must be set with dbpaths and be of equal length")
	}

	for i := range pathsToDBs {
		pathsToDBs[i] = strings.TrimRight(pathsToDBs[i], "/") + "/"
	}

	return nil
}
