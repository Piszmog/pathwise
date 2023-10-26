package db

import (
	"database/sql"
	"fmt"
	_ "github.com/libsql/libsql-client-go/libsql"
)

type Database interface {
	DB() *sql.DB
	Close() error
}

func New(dbType DatabaseType, opts DatabaseOpts) (Database, error) {
	switch dbType {
	case DatabaseTypeFile:
		return newFileDB(opts.URL)
	case DatabaseTypeTurso:
		return newTursoDB(opts.URL, opts.Token)
	case DatabaseTypeURL:
		return newURLDB(opts.URL)
	default:
		return nil, fmt.Errorf("unknown database type: %s", dbType)
	}
}

type DatabaseType string

const (
	DatabaseTypeFile  DatabaseType = "file"
	DatabaseTypeTurso DatabaseType = "turso"
	DatabaseTypeURL   DatabaseType = "url"
)

type DatabaseOpts struct {
	URL   string
	Token string
}

func Init(database Database) error {
	_, err := database.DB().Exec(
		`CREATE TABLE IF NOT EXISTS job_applications (
            id INTEGER PRIMARY KEY,
            company TEXT NOT NULL,
            title TEXT NOT NULL,
            url TEXT,
            status TEXT NOT NULL DEFAULT 'applied',
            applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        )`,
	)
	if err != nil {
		return err
	}

	_, err = database.DB().Exec(
		`CREATE TABLE IF NOT EXISTS job_application_notes (
            id INTEGER PRIMARY KEY,
            job_application_id INTEGER NOT NULL,
            note TEXT NOT NULL,
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (job_application_id) REFERENCES job_applications(id) ON DELETE CASCADE
        )`,
	)
	if err != nil {
		return err
	}

	_, err = database.DB().Exec(
		`CREATE TABLE IF NOT EXISTS job_application_status_histories (
            id INTEGER PRIMARY KEY,
            job_application_id INTEGER NOT NULL,
            status TEXT NOT NULL DEFAULT 'applied',
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (job_application_id) REFERENCES job_applications(id) ON DELETE CASCADE
        )`,
	)
	if err != nil {
		return err
	}
	return nil
}
