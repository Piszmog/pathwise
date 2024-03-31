package main

import (
	"os"

	"github.com/Piszmog/pathwise/auth"
	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/logger"
	"github.com/Piszmog/pathwise/server"
	"github.com/Piszmog/pathwise/server/router"
	"github.com/Piszmog/pathwise/version"
)

func main() {
	l := logger.New(os.Getenv("LOG_LEVEL"))

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "./db.sqlite3"
	}

	database, err := db.New(
		l,
		db.DatabaseOpts{URL: dbURL, Token: os.Getenv("DB_TOKEN")},
	)

	if err != nil {
		l.Error("failed to create database", "error", err)
		return
	}
	defer database.Close()

	if err = db.Init(database); err != nil {
		l.Error("failed to initialize database", "error", err)
		return
	}

	v := os.Getenv("VERSION")
	if v != "" {
		version.Value = v
	}

	sessionStore := &store.SessionStore{Database: database}
	sessionJanitor := auth.SessionJanitor{
		Logger: l,
		Store:  sessionStore,
	}
	go sessionJanitor.Run()

	r := router.New(l, database, sessionStore)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server.New(l, ":"+port, server.WithHandler(r)).StartAndWait()
}
