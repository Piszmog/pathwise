package db

import (
	"database/sql"
	"log/slog"

	_ "modernc.org/sqlite"
)

type FileDB struct {
	logger *slog.Logger
	db     *sql.DB
}

func (d *FileDB) DB() *sql.DB {
	return d.db
}

func (d *FileDB) Logger() *slog.Logger {
	return d.logger
}

func (d *FileDB) Close() error {
	return d.db.Close()
}

func newFileDB(logger *slog.Logger, path string) (*FileDB, error) {
	db, err := sql.Open("libsql", "file:"+path)
	if err != nil {
		return nil, err
	}
	return &FileDB{logger: logger, db: db}, nil
}
