package store

import (
	"context"
	"database/sql"

	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/types"
)

type UserStore struct {
	Database db.Database
}

func (s *UserStore) Insert(ctx context.Context, user types.User) error {
	_, err := s.Database.DB().ExecContext(ctx, userInsertQuery, user.Email, user.Password)
	return err
}

const userInsertQuery = `INSERT INTO users (email, password) VALUES (?, ?)`

func (s *UserStore) GetByEmail(ctx context.Context, email string) (types.User, error) {
	row := s.Database.DB().QueryRowContext(ctx, userGetQuery, email)
	return scanUser(row)
}

const userGetQuery = `SELECT id, email, password FROM users WHERE email = ?`

func (s *UserStore) GetByID(ctx context.Context, id int64) (types.User, error) {
	row := s.Database.DB().QueryRowContext(ctx, userGetByIDQuery, id)
	return scanUser(row)
}

const userGetByIDQuery = `SELECT id, email, password FROM users WHERE id = ?`

func scanUser(row *sql.Row) (types.User, error) {
	var user types.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
	)
	return user, err
}

func (s *UserStore) Delete(ctx context.Context, id int64) error {
	_, err := s.Database.DB().ExecContext(ctx, userDeleteQuery, id)
	return err
}

const userDeleteQuery = `DELETE FROM users WHERE id = ?`

func (s *UserStore) UpdatePassword(ctx context.Context, id int64, password string) error {
	_, err := s.Database.DB().ExecContext(ctx, userUpdatePasswordQuery, password, id)
	return err
}

const userUpdatePasswordQuery = `UPDATE users SET password = ? WHERE id = ?`
