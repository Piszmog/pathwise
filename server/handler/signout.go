package handler

import (
	"net/http"
	"time"
)

func (h *Handler) Signout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		h.Logger.Error("failed to get session cookie", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = h.Database.Queries().DeleteSessionByToken(r.Context(), cookie.Value); err != nil {
		h.Logger.Error("failed to delete session", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}
