package db

import (
	"database/sql"
	"log/slog"

	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/tursodatabase/go-libsql"
)

type EmbeddedReplicaDB struct {
	logger    *slog.Logger
	connector *libsql.Connector
	db        *sql.DB
	queries   *queries.Queries
}

var _ Database = (*EmbeddedReplicaDB)(nil)

func (d *EmbeddedReplicaDB) DB() *sql.DB {
	return d.db
}

func (d *EmbeddedReplicaDB) Queries() *queries.Queries {
	return d.queries
}

func (d *EmbeddedReplicaDB) Logger() *slog.Logger {
	return d.logger
}

func (d *EmbeddedReplicaDB) Close() error {
	if err := d.connector.Close(); err != nil {
		return err
	}
	return d.db.Close()
}

func newEmbeddedReplicaDB(logger *slog.Logger, dbOpts DatabaseOpts) (*EmbeddedReplicaDB, error) {
	var opts []libsql.Option
	opts = append(opts, libsql.WithAuthToken(dbOpts.Token))
	if dbOpts.EncryptionKey != "" {
		opts = append(opts, libsql.WithEncryption(dbOpts.EncryptionKey))
	}
	connector, err := libsql.NewEmbeddedReplicaConnector(
		dbOpts.URL,
		dbOpts.PrimaryURL,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(connector)
	return &EmbeddedReplicaDB{
		logger:    logger,
		connector: connector,
		db:        db,
		queries:   queries.New(db),
	}, nil
}
