package db

import (
	"database/sql"
	"log/slog"
)

type Database interface {
	DB() *sql.DB
	Logger() *slog.Logger
	Close() error
}

func New(logger *slog.Logger, opts DatabaseOpts) (Database, error) {
	if opts.Token == "" {
		return newLocalDB(logger, opts.URL)
	} else {
		return newRemoteDB(logger, opts.URL, opts.Token)
	}
}

type DatabaseOpts struct {
	URL   string
	Token string
}

func Init(database Database) error {
	_, err := database.DB().Exec(`PRAGMA foreign_keys = ON`)
	if err != nil {
		return err
	}

	_, err = database.DB().Exec(
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	)
	if err != nil {
		return err
	}

	_, err = database.DB().Exec(
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			token TEXT NOT NULL UNIQUE,
			user_agent TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	)
	if err != nil {
		return err
	}

	_, err = database.DB().Exec(
		`CREATE TABLE IF NOT EXISTS job_applications (
            id INTEGER PRIMARY KEY,
            company TEXT NOT NULL,
            title TEXT NOT NULL,
            url TEXT,
            status TEXT NOT NULL DEFAULT 'applied',
            applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        )`,
	)
	if err != nil {
		return err
	}

	_, err = database.DB().Exec(
		`CREATE INDEX IF NOT EXISTS job_applications_user_id_idx ON job_applications(user_id)`,
	)
	if err != nil {
		return err
	}

	_, err = database.DB().Exec(
		`CREATE INDEX IF NOT EXISTS job_applications_stats_idx ON job_applications(user_id, id, company, status, applied_at)`,
	)
	if err != nil {
		return err
	}

	_, err = database.DB().Exec(
		`CREATE INDEX IF NOT EXISTS job_applications_user_id_updated_at_idx ON job_applications(user_id, updated_at)`,
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

	_, err = database.DB().Exec(
		`CREATE INDEX IF NOT EXISTS status_job_application_id_created_at_idx ON job_application_status_histories(status, job_application_id, created_at)`,
	)
	if err != nil {
		return err
	}

	return nil
}
