package db

import (
	"database/sql"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"log/slog"
)

type RemoteDB struct {
	logger *slog.Logger
	db     *sql.DB
}

func (d *RemoteDB) DB() *sql.DB {
	return d.db
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
	return &RemoteDB{logger: logger, db: db}, nil
}
