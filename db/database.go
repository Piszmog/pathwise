package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/Piszmog/pathwise/db/queries"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Database interface {
	DB() *sql.DB
	Queries() *queries.Queries
	Logger() *slog.Logger
	Close() error
}

func New(logger *slog.Logger, opts DatabaseOpts) (Database, error) {
	if opts.Token == "" {
		return newLocalDB(logger, opts.URL)
	} else {
		return newRemoteDB(logger, opts.URL, opts.Token)
	}
}

type DatabaseOpts struct {
	URL   string
	Token string
}

func Migrate(db Database) error {
	driver, err := sqlite3.WithInstance(db.DB(), &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	iofsDriver, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs: %w", err)
	}
	defer iofsDriver.Close()

	m, err := migrate.NewWithInstance("iofs", iofsDriver, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	return m.Up()
}
