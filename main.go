package main

import (
	"embed"
	"os"

	"github.com/Piszmog/pathwise/auth"
	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/logger"
	"github.com/Piszmog/pathwise/server"
	"github.com/Piszmog/pathwise/server/router"
)

//go:embed assets/*
var assets embed.FS

func main() {
	l := logger.New(os.Getenv("LOG_LEVEL"))

	databaseType := getDatabaseType()

	var database db.Database
	var err error
	switch databaseType {
	case db.DatabaseTypeFile:
		l.Info("using file database")
		database, err = db.New(db.DatabaseTypeFile, db.DatabaseOpts{URL: "./db.sqlite3"})
	case db.DatabaseTypeTurso:
		l.Info("using turso database")
		database, err = db.New(
			db.DatabaseTypeTurso,
			db.DatabaseOpts{URL: os.Getenv("DB_URL"), Token: os.Getenv("DB_TOKEN")},
		)
	default:
		l.Error("unknown database type", "type", databaseType)
		return
	}

	if err != nil {
		l.Error("failed to create database", "error", err)
		return
	}
	defer database.Close()

	if err = db.Init(database); err != nil {
		l.Error("failed to initialize database", "error", err)
		return
	}

	sessionStore := &store.SessionStore{Database: database}
	sessionJanitor := auth.SessionJanitor{
		Logger: l,
		Store:  sessionStore,
	}
	go sessionJanitor.Run()

	r := router.New(l, database, assets, sessionStore)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server.New(l, ":"+port, server.WithHandler(r)).StartAndWait()
}

func getDatabaseType() db.DatabaseType {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		return db.DatabaseTypeFile
	}
	return db.DatabaseType(dbType)
}
