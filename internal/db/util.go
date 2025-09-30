package db

import "database/sql"

func NewNullString(value string) sql.NullString {
	var isValid bool
	if value != "" {
		isValid = true
	}
	return sql.NullString{
		Valid:  isValid,
		String: value,
	}
}
