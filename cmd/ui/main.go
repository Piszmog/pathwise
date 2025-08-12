package main

import (
	"os"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/logger"
	"github.com/Piszmog/pathwise/internal/version"
	"github.com/Piszmog/pathwise/ui/server"
	"github.com/Piszmog/pathwise/ui/server/router"
)

func main() {
	l := logger.New(os.Getenv("LOG_LEVEL"), os.Getenv("LOG_OUTPUT"))

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "./db.sqlite3"
	}

	l.Info("opts", "db",
		db.DatabaseOpts{
			URL:           dbURL,
			PrimaryURL:    os.Getenv("DB_URL_PRIMARY"),
			Token:         os.Getenv("DB_TOKEN"),
			EncryptionKey: os.Getenv("ENC_KEY"),
		},
	)

	database, err := db.New(
		l,
		db.DatabaseOpts{
			URL:           dbURL,
			PrimaryURL:    os.Getenv("DB_URL_PRIMARY"),
			Token:         os.Getenv("DB_TOKEN"),
			EncryptionKey: os.Getenv("ENC_KEY"),
		},
	)
	if err != nil {
		l.Error("failed to create database", "error", err)
		return
	}
	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			l.Error("failed to close database", "error", closeErr)
		}
	}()

	if _, err = database.DB().Exec("PRAGMA foreign_keys = ON;"); err != nil {
		l.Error("failed to enable foreign keys", "error", err)
		return
	}

	// if err = db.Migrate(database); err != nil && !errors.Is(err, migrate.ErrNoChange) {
	// 	l.Error("failed to migrate database", "error", err)
	// 	return
	// }

	v := os.Getenv("VERSION")
	if v != "" {
		version.Value = v
	}

	r := router.New(l, database)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server.New(l, ":"+port, server.WithHandler(r)).StartAndWait()
}
