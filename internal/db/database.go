package db

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/Piszmog/pathwise/internal/db/queries"
)

type Database interface {
	DB() *sql.DB
	Queries() *queries.Queries
	Logger() *slog.Logger
	Close() error
}

func New(logger *slog.Logger, opts DatabaseOpts) (Database, error) {
	var db Database
	var err error
	if opts.SyncURL != "" && opts.Token != "" {
		db, err = newEmbeddedDB(logger, opts)
	} else if opts.Token != "" {
		db, err = newRemoteDB(logger, opts)
	} else {
		db, err = newLocalDB(logger, opts)
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
	Token         string
	SyncURL       string
	EncryptionKey string
	SyncInterval  time.Duration
}
