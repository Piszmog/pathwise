package db

import (
	"database/sql"
	"log/slog"

	"github.com/Piszmog/pathwise/internal/db/queries"
	_ "github.com/tursodatabase/go-libsql"
)

type RemoteDB struct {
	logger  *slog.Logger
	db      *sql.DB
	queries *queries.Queries
}

var _ Database = (*RemoteDB)(nil)

func (d *RemoteDB) DB() *sql.DB {
	return d.db
}

func (d *RemoteDB) Queries() *queries.Queries {
	return d.queries
}

func (d *RemoteDB) Logger() *slog.Logger {
	return d.logger
}

func (d *RemoteDB) Close() error {
	return d.db.Close()
}

func newRemoteDB(logger *slog.Logger, opts DatabaseOpts) (*RemoteDB, error) {
	fullURL := "libsql://" + opts.URL + "?authToken=" + opts.Token
	db, err := sql.Open("libsql", fullURL)
	if err != nil {
		return nil, err
	}
	return &RemoteDB{logger: logger, db: db, queries: queries.New(db)}, nil
}
