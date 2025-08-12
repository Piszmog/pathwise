package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
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
	var db Database
	var err error
	if opts.Token == "" {
		db, err = newLocalDB(logger, opts)
	} else if opts.PrimaryURL == "" {
		db, err = newRemoteDB(logger, opts)
	} else {
		db, err = newEmbeddedReplicaDB(logger, opts)
	}
	if err != nil {
		return nil, err
	}
	if err = db.DB().PingContext(context.Background()); err != nil {
		return nil, err
	}
	return db, nil
}

type DatabaseOpts struct {
	URL           string
	PrimaryURL    string
	Token         string
	EncryptionKey string
}

func Migrate(db Database) error {
	driver, err := sqlite.WithInstance(db.DB(), &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	iofsDriver, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs: %w", err)
	}
	defer func() {
		_ = iofsDriver.Close()
	}()

	m, err := migrate.NewWithInstance("iofs", iofsDriver, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration: %w", err)
	}

	return m.Up()
}
