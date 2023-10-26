package db

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

type FileDB struct {
	db *sql.DB
}

func (d *FileDB) DB() *sql.DB {
	return d.db
}

func (d *FileDB) Close() error {
	return d.db.Close()
}

func newFileDB(path string) (*FileDB, error) {
	db, err := sql.Open("libsql", "file:"+path)
	if err != nil {
		return nil, err
	}
	return &FileDB{db: db}, nil
}
