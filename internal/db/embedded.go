package db

import (
	"database/sql"
	"log/slog"

	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/tursodatabase/go-libsql"
)

type EmbeddedDB struct {
	logger    *slog.Logger
	connector *libsql.Connector
	db        *sql.DB
	queries   *queries.Queries
}

var _ Database = (*EmbeddedDB)(nil)

func (d *EmbeddedDB) DB() *sql.DB {
	return d.db
}

func (d *EmbeddedDB) Queries() *queries.Queries {
	return d.queries
}

func (d *EmbeddedDB) Logger() *slog.Logger {
	return d.logger
}

func (d *EmbeddedDB) Close() error {
	if err := d.db.Close(); err != nil {
		return err
	}
	return d.connector.Close()
}

func newEmbeddedDB(logger *slog.Logger, opts DatabaseOpts) (*EmbeddedDB, error) {
	connectorOpts := []libsql.Option{
		libsql.WithAuthToken(opts.Token),
	}

	if opts.EncryptionKey != "" {
		connectorOpts = append(connectorOpts, libsql.WithEncryption(opts.EncryptionKey))
	}

	if opts.SyncInterval > 0 {
		connectorOpts = append(connectorOpts, libsql.WithSyncInterval(opts.SyncInterval))
	}

	connector, err := libsql.NewEmbeddedReplicaConnector(
		opts.URL,
		opts.SyncURL,
		connectorOpts...,
	)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)
	return &EmbeddedDB{
		logger:    logger,
		connector: connector,
		db:        db,
		queries:   queries.New(db),
	}, nil
}
