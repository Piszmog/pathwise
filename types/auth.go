package types

import "time"

type User struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
	Password  string
	ID        int
}

type Session struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
	Token     string
	UserID    int
	ID        int
}
