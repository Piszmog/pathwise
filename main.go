package main

import (
	"embed"
	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/logger"
	"github.com/Piszmog/pathwise/server"
	"github.com/Piszmog/pathwise/server/router"
	"os"
)

//go:embed assets/*
var assets embed.FS

func main() {
	l := logger.New(os.Getenv("LOG_LEVEL"))

	database, err := db.New(db.DatabaseTypeFile, db.DatabaseOpts{URL: "./db.sqlite3"})
	if err != nil {
		l.Error("failed to create database", "error", err)
		return
	}
	defer database.Close()

	if err = db.Init(database); err != nil {
		l.Error("failed to initialize database", "error", err)
		return
	}

	r := router.New(l, database, assets)

	server.New(l, ":8080", server.WithHandler(r)).StartAndWait()
}
