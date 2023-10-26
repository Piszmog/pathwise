package db

import "database/sql"

type TursoDB struct {
	db *sql.DB
}

func (d *TursoDB) DB() *sql.DB {
	return d.db
}

func (d *TursoDB) Close() error {
	return d.db.Close()
}

func newTursoDB(name string, token string) (*TursoDB, error) {
	db, err := sql.Open("libsql", "libsql://"+name+".turso.io?authToken="+token)
	if err != nil {
		return nil, err
	}
	return &TursoDB{db: db}, nil
}
