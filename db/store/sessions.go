package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/types"
)

type SessionStore struct {
	Database db.Database
}

func (s *SessionStore) Insert(ctx context.Context, session types.Session) error {
	if session.UserID == 0 {
		return fmt.Errorf("user id cannot be 0")
	} else if session.Token == "" {
		return fmt.Errorf("token cannot be empty")
	} else if session.ExpiresAt.IsZero() {
		return fmt.Errorf("expires at cannot be zero")
	} else if session.UserAgent == "" {
		return fmt.Errorf("user agent cannot be empty")
	}
	_, err := s.Database.DB().ExecContext(ctx, sessionInsertQuery, session.UserID, session.UserAgent, session.Token, session.ExpiresAt)
	return err
}

const sessionInsertQuery = `INSERT INTO sessions (user_id, user_agent, token, expires_at) VALUES (?, ?, ?, ?)`

func (s *SessionStore) Get(ctx context.Context, token string) (types.Session, error) {
	row := s.Database.DB().QueryRowContext(ctx, sessionGetQuery, token)
	return scanSession(row)
}

const sessionGetQuery = `SELECT created_at, expires_at, token, user_id FROM sessions WHERE token = ?`

func scanSession(row *sql.Row) (types.Session, error) {
	var session types.Session
	err := row.Scan(
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.Token,
		&session.UserID,
	)
	return session, err
}

func (s *SessionStore) DeleteByUserID(ctx context.Context, userId int) error {
	_, err := s.Database.DB().ExecContext(ctx, sessionDeleteQuery, userId)
	return err
}

const sessionDeleteQuery = `DELETE FROM sessions WHERE user_id = ?`

func (s *SessionStore) DeleteByToken(ctx context.Context, token string) error {
	_, err := s.Database.DB().ExecContext(ctx, sessionDeleteByTokenQuery, token)
	return err
}

const sessionDeleteByTokenQuery = `DELETE FROM sessions WHERE token = ?`

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
