package db

import (
	"database/sql"
	"log/slog"
)

type URLDB struct {
	logger *slog.Logger
	db     *sql.DB
}

func (d *URLDB) DB() *sql.DB {
	return d.db
}

func (d *URLDB) Logger() *slog.Logger {
	return d.logger
}

func (d *URLDB) Close() error {
	return d.db.Close()
}

func newURLDB(logger *slog.Logger, url string) (*URLDB, error) {
	db, err := sql.Open("libsql", url)
	if err != nil {
		return nil, err
	}
	return &URLDB{logger: logger, db: db}, nil
}
