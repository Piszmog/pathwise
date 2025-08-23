package db

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/Piszmog/pathwise/internal/db/queries"
	_ "github.com/tursodatabase/go-libsql"
)

type LocalDB struct {
	logger  *slog.Logger
	db      *sql.DB
	queries *queries.Queries
}

var _ Database = (*LocalDB)(nil)

func (d *LocalDB) DB() *sql.DB {
	return d.db
}

func (d *LocalDB) Queries() *queries.Queries {
	return d.queries
}

func (d *LocalDB) Logger() *slog.Logger {
	return d.logger
}

func (d *LocalDB) Close() error {
	return d.db.Close()
}

func newLocalDB(logger *slog.Logger, opts DatabaseOpts) (*LocalDB, error) {
	db, err := sql.Open("libsql", "file:"+opts.URL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)
	return &LocalDB{logger: logger, db: db, queries: queries.New(db)}, nil
}
