package middleware

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Piszmog/pathwise/db"
	"github.com/Piszmog/pathwise/db/queries"
)

type AuthMiddleware struct {
	Logger   *slog.Logger
	Database db.Database
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isHxRequest := r.Header.Get("HX-Request") == "true"
		cookie, err := r.Cookie("session")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				m.Logger.WarnContext(r.Context(), "no session cookie", "err", err)
				w.Header().Set("HX-Redirect", "/signin")
				if !isHxRequest {
					http.Redirect(w, r, "/signin", http.StatusSeeOther)
				}
				return
			}
			m.Logger.ErrorContext(r.Context(), "failed to get session cookie", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		session, err := m.Database.Queries().GetSessionByToken(r.Context(), cookie.Value)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				m.Logger.ErrorContext(r.Context(), "failed to get session", "err", err)
			}
			w.Header().Set("HX-Redirect", "/signin")
			if !isHxRequest {
				http.Redirect(w, r, "/signin", http.StatusSeeOther)
			}
			return
		}

		if session.ExpiresAt.Before(time.Now()) {
			m.Logger.DebugContext(r.Context(), "session expired", "session", session)
			err = m.Database.Queries().DeleteSessionByToken(r.Context(), cookie.Value)
			if err != nil {
				m.Logger.ErrorContext(r.Context(), "failed to delete session", "err", err)
			}
			w.Header().Set("HX-Redirect", "/signin")
			if !isHxRequest {
				http.Redirect(w, r, "/signin", http.StatusSeeOther)
			}
			return
		}

		if session.ExpiresAt.Sub(session.CreatedAt) < 5 {
			m.Logger.DebugContext(r.Context(), "refreshing session", "session", session)
			err = m.Database.Queries().UpdateSessionExpiresAt(r.Context(), queries.UpdateSessionExpiresAtParams{Token: session.Token, ExpiresAt: session.ExpiresAt.Add(24 * 7)})
			if err != nil {
				m.Logger.ErrorContext(r.Context(), "failed to refresh session", "err", err)
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

		r.Header.Set("USER-ID", strconv.FormatInt(session.UserID, 10))

		next.ServeHTTP(w, r)
	})
}
