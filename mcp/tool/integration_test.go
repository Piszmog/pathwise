//go:build integration

package tool_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup any remaining test database files
	cleanupLeftoverTestFiles()

	os.Exit(code)
}

func cleanupLeftoverTestFiles() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	matches, err := filepath.Glob(filepath.Join(wd, "integration-test-*.sqlite3"))
	if err != nil {
		return
	}

	for _, file := range matches {
		_ = os.Remove(file)
	}
}

func runMigrations(t *testing.T, dbFile string) error {
	t.Helper()

	repoRoot, err := getRepoRoot()
	if err != nil {
		return fmt.Errorf("could not find repo root: %v", err)
	}

	migrateScript := filepath.Join(repoRoot, "migrate.sh")
	cmd := exec.Command(migrateScript, "-p", "sqlite3", "-u", dbFile, "-d", "up")
	cmd.Dir = repoRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If migration fails due to dirty state, try to force it
		if strings.Contains(string(output), "Dirty database version") {
			t.Logf("Database in dirty state, attempting to force migration")

			// Try to force the version and re-run
			forceCmd := exec.Command("go", "run", "-tags", "sqlite3",
				"github.com/golang-migrate/migrate/v4/cmd/migrate",
				"-source", "file://./internal/db/migrations",
				"-database", fmt.Sprintf("sqlite3://%s", dbFile),
				"force", "20250804014440") // Latest migration version
			forceCmd.Dir = repoRoot

			if forceOutput, forceErr := forceCmd.CombinedOutput(); forceErr != nil {
				return fmt.Errorf("force migration failed: %v, output: %s", forceErr, forceOutput)
			}

			// Try migration again
			retryOutput, retryErr := cmd.CombinedOutput()
			if retryErr != nil {
				return fmt.Errorf("retry migration failed: %v, output: %s", retryErr, retryOutput)
			}
		} else {
			return fmt.Errorf("migration failed: %v, output: %s", err, output)
		}
	}

	return nil
}

func getRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}

	return "", fmt.Errorf("could not find go.mod")
}

func setupTestDB(t *testing.T) db.Database {
	t.Helper()

	// Create truly unique filename with test name and timestamp
	dbFileName := fmt.Sprintf("integration-test-%s-%d-%d.sqlite3",
		strings.ReplaceAll(t.Name(), "/", "_"),
		time.Now().UnixNano(),
		rand.Int())

	// Get absolute path for database file
	wd, err := os.Getwd()
	require.NoError(t, err)
	dbFile := filepath.Join(wd, dbFileName)

	// Ensure cleanup happens even if setup fails
	var database db.Database
	cleanup := func() {
		if database != nil {
			if err := database.Close(); err != nil {
				t.Logf("Failed to close database: %v", err)
			}
		}

		// Wait a moment for file handles to be released
		time.Sleep(10 * time.Millisecond)

		// Force remove file (ignore errors if already deleted)
		if err := os.Remove(dbFile); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove database file %s: %v", dbFile, err)
		}
	}
	t.Cleanup(cleanup)

	database, err = db.New(setupTestLogger(), db.DatabaseOpts{URL: dbFile})
	require.NoError(t, err)

	ctx := context.Background()
	_, err = database.DB().ExecContext(ctx, "PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	err = runMigrations(t, dbFile)
	require.NoError(t, err)

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
