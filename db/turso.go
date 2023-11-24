package db

import (
	"database/sql"
	"log/slog"
)

type TursoDB struct {
	logger *slog.Logger
	db     *sql.DB
}

func (d *TursoDB) DB() *sql.DB {
	return d.db
}

func (d *TursoDB) Logger() *slog.Logger {
	return d.logger
}

func (d *TursoDB) Close() error {
	return d.db.Close()
}

func newTursoDB(logger *slog.Logger, name string, token string) (*TursoDB, error) {
	db, err := sql.Open("libsql", "libsql://"+name+".turso.io?authToken="+token)
	if err != nil {
		return nil, err
	}
	return &TursoDB{logger: logger, db: db}, nil
}
