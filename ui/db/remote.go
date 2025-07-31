package db

import (
	"database/sql"
	"log/slog"

	"github.com/Piszmog/pathwise/ui/db/queries"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
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

func newRemoteDB(logger *slog.Logger, name string, token string) (*RemoteDB, error) {
	db, err := sql.Open("libsql", "libsql://"+name+".turso.io?authToken="+token)
	if err != nil {
		return nil, err
	}
	return &RemoteDB{logger: logger, db: db, queries: queries.New(db)}, nil
}
