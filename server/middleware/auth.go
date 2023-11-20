package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Piszmog/pathwise/db/store"
)

type AuthMiddleware struct {
	Logger       *slog.Logger
	SessionStore *store.SessionStore
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isHxRequest := r.Header.Get("HX-Request") == "true"
		cookie, err := r.Cookie("session")
		if err != nil {
			if err == http.ErrNoCookie {
				m.Logger.Error("no session cookie", "err", err)
				w.Header().Set("HX-Redirect", "/signin")
				if !isHxRequest {
					http.Redirect(w, r, "/signin", http.StatusSeeOther)
				}
				return
			}
			m.Logger.Error("failed to get session cookie", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		session, err := m.SessionStore.Get(r.Context(), cookie.Value)
		if err != nil {
			m.Logger.Error("failed to get session", "err", err)
			w.Header().Set("HX-Redirect", "/signin")
			if !isHxRequest {
				http.Redirect(w, r, "/signin", http.StatusSeeOther)
			}
			return
		}

		if session.ExpiresAt.Before(time.Now()) {
			m.Logger.Debug("session expired", "session", session)
			err = m.SessionStore.DeleteByToken(r.Context(), cookie.Value)
			if err != nil {
				m.Logger.Error("failed to delete session", "err", err)
			}
			w.Header().Set("HX-Redirect", "/signin")
			if !isHxRequest {
				http.Redirect(w, r, "/signin", http.StatusSeeOther)
			}
			return
		}

		if session.ExpiresAt.Sub(session.CreatedAt) < 5 {
			m.Logger.Debug("refreshing session", "session", session)
			err = m.SessionStore.Refresh(r.Context(), session.Token, session.ExpiresAt.Add(24*7))
			if err != nil {
				m.Logger.Error("failed to refresh session", "err", err)
			} else {
				http.SetCookie(w, &http.Cookie{
					Name:     "session",
					Value:    session.Token,
					Expires:  session.ExpiresAt,
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
				})
			}
		}

		r.Header.Set("USER-ID", strconv.Itoa(session.UserID))

		next.ServeHTTP(w, r)
	})
}
