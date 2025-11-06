package testutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/tursodatabase/go-libsql"
)

// RunMigrations executes all migration files in order against the given database file.
// This is a test-only utility that reads .up.sql files directly without using external tools.
func RunMigrations(dbFile string) error {
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	db, err := sql.Open("libsql", "file:"+dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	for _, file := range migrationFiles {
		if err := executeMigrationFile(db, file); err != nil {
			return fmt.Errorf("migration %s failed: %w", filepath.Base(file), err)
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
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
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
func executeMigrationFile(db *sql.DB, path string) error {
	cleanPath := filepath.Clean(path)
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Split SQL statements and execute them individually for go-libsql compatibility
	statements := strings.SplitSeq(string(content), ";")
	for stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if _, err := db.ExecContext(context.Background(), stmt); err != nil {
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}
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

	return "", errGoModNotFound
}

var errGoModNotFound = errors.New("could not find go.mod file")
