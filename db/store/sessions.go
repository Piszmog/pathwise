package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/types"
)

type SessionStore struct {
	Database db.Database
}

func (s *SessionStore) Insert(ctx context.Context, session types.Session) error {
	_, err := s.Database.DB().ExecContext(ctx, sessionInsertQuery, session.UserID, session.Token, session.ExpiresAt)
	return err
}

const sessionInsertQuery = `INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)`

func (s *SessionStore) Get(ctx context.Context, token string) (types.Session, error) {
	row := s.Database.DB().QueryRowContext(ctx, sessionGetQuery, token)
	return scanSession(row)
}

const sessionGetQuery = `SELECT id, created_at, updated_at, expires_at, token, user_id FROM sessions WHERE token = ?`

func scanSession(row *sql.Row) (types.Session, error) {
	var session types.Session
	err := row.Scan(
		&session.ID,
		&session.CreatedAt,
		&session.UpdatedAt,
		&session.ExpiresAt,
		&session.Token,
		&session.UserID,
	)
	return session, err
}

func (s *SessionStore) Delete(ctx context.Context, userId int) error {
	_, err := s.Database.DB().ExecContext(ctx, sessionDeleteQuery, userId)
	return err
}

const sessionDeleteQuery = `DELETE FROM sessions WHERE user_id = ?`

func (s *SessionStore) DeleteExpired(ctx context.Context) error {
	_, err := s.Database.DB().ExecContext(ctx, sessionDeleteExpiredQuery, time.Now())
	return err
}

const sessionDeleteExpiredQuery = `DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`

func (s *SessionStore) Refresh(ctx context.Context, token string, expiresAt time.Time) error {
	_, err := s.Database.DB().ExecContext(ctx, sessionRefreshQuery, expiresAt, token)
	return err
}

const sessionRefreshQuery = `UPDATE sessions SET expires_at = ? WHERE token = ?`
