//go:build integration

package tool_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) db.Database {
	t.Helper()

	dbFile := fmt.Sprintf("integration-test-%d.sqlite3", time.Now().UnixNano())

	database, err := db.New(setupTestLogger(), db.DatabaseOpts{URL: dbFile})
	require.NoError(t, err)

	ctx := context.Background()
	_, err = database.DB().ExecContext(ctx, "PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	err = db.Migrate(database)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = database.Close()
		_ = os.Remove(dbFile)
	})

	return database
}

func cleanupTestDB(t *testing.T, database db.Database) {
	t.Helper()
	if database != nil {
		_ = database.Close()
	}
}

func createTestUser(t *testing.T, db *sql.DB, userID int64) {
	t.Helper()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() {
		_ = tx.Rollback()
	}()

	email := fmt.Sprintf("test-user-%d@example.com", userID)
	//nolint:gosec
	hashedPassword := "$2a$14$YRpu0/fntbFMA8Zne3hyLufuYhNkeoM/.68SvNXduN0/eE/s0A3hm"

	_, err = tx.ExecContext(ctx,
		"INSERT INTO users (id, email, password, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		userID, email, hashedPassword,
	)
	require.NoError(t, err)

	_, err = tx.ExecContext(ctx,
		"INSERT INTO job_application_stats (user_id, total_applications, total_companies, total_applied) VALUES (?, 0, 0, 0)",
		userID,
	)
	require.NoError(t, err)

	require.NoError(t, tx.Commit())
}

func insertJobApplication(t *testing.T, db *sql.DB, userID int64, company, title, status string) int64 {
	t.Helper()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() {
		_ = tx.Rollback()
	}()

	var jobID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO job_applications (
			user_id, company, title, url, status, applied_at, archived,
			salary_min, salary_max, salary_currency, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, 0, NULL, NULL, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`, userID, company, title, fmt.Sprintf("https://%s.com/jobs", strings.ToLower(strings.ReplaceAll(company, " ", ""))), status).Scan(&jobID)
	require.NoError(t, err)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO job_application_status_histories (
			job_application_id, status, created_at
		) VALUES (?, ?, CURRENT_TIMESTAMP)
	`, jobID, status)
	require.NoError(t, err)

	require.NoError(t, tx.Commit())
	return jobID
}

func insertJobApplicationNote(t *testing.T, db *sql.DB, jobApplicationID int64, note string) int64 {
	t.Helper()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() {
		_ = tx.Rollback()
	}()

	var noteID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO job_application_notes (job_application_id, note, created_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		RETURNING id
	`, jobApplicationID, note).Scan(&noteID)
	require.NoError(t, err)

	require.NoError(t, tx.Commit())
	return noteID
}

func insertJobApplicationStatusHistory(t *testing.T, db *sql.DB, jobApplicationID int64, status string) int64 {
	t.Helper()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() {
		_ = tx.Rollback()
	}()

	var historyID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO job_application_status_histories (job_application_id, status, created_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		RETURNING id
	`, jobApplicationID, status).Scan(&historyID)
	require.NoError(t, err)

	require.NoError(t, tx.Commit())
	return historyID
}

func setupTestLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
