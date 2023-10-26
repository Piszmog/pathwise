package db

import "database/sql"

type URLDB struct {
	db *sql.DB
}

func (d *URLDB) DB() *sql.DB {
	return d.db
}

func (d *URLDB) Close() error {
	return d.db.Close()
}

func newURLDB(url string) (*URLDB, error) {
	db, err := sql.Open("libsql", url)
	if err != nil {
		return nil, err
	}
	return &URLDB{db: db}, nil
}
