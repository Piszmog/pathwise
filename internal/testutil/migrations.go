package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// RunMigrations executes all migration files in order against the given database file.
// This is a test-only utility that reads .up.sql files directly without using external tools.
func RunMigrations(t *testing.T, dbFile string) error {
	t.Helper()

	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %v", err)
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Execute each migration file in order
	for _, file := range migrationFiles {
		if err := executeMigrationFile(db, file); err != nil {
			return fmt.Errorf("migration %s failed: %v", filepath.Base(file), err)
		}
	}

	return nil
}

// getMigrationFiles returns all .up.sql migration files in sorted order
func getMigrationFiles() ([]string, error) {
	repoRoot, err := getRepoRoot()
	if err != nil {
		return nil, err
	}

	migrationsDir := filepath.Join(repoRoot, "internal", "db", "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %v", err)
	}

	var migrationFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			migrationFiles = append(migrationFiles, filepath.Join(migrationsDir, name))
		}
	}

	// Sort files to ensure proper migration order
	sort.Strings(migrationFiles)

	return migrationFiles, nil
}

// executeMigrationFile reads and executes a single migration file
func executeMigrationFile(db *sql.DB, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %v", err)
	}

	// Execute the SQL content
	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %v", err)
	}

	return nil
}

// getRepoRoot finds the project root directory by looking for go.mod
func getRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for go.mod
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

	return "", fmt.Errorf("could not find go.mod file")
}
