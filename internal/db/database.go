package db

import (
	"context"
	"database/sql"
	"log/slog"

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
	if opts.Token == "" {
		db, err = newLocalDB(logger, opts.URL)
	} else {
		db, err = newRemoteDB(logger, opts.URL, opts.Token)
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
	URL   string
	Token string
}
