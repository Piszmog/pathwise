package utils

import (
	"net/http"
	"os"
	"time"
)

func IsProduction() bool {
	env := os.Getenv("ENV")
	return env == "production" || env == "prod"
}

func SetSessionCookie(w http.ResponseWriter, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    value,
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Secure:   IsProduction(),
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	SetSessionCookie(w, "", time.Now().Add(-1*time.Hour))
}
