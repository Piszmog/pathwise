package handler

import (
	"net/http"

	"github.com/Piszmog/pathwise/internal/ui/utils"
)

func (h *Handler) Signout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get session cookie", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = h.Database.Queries().DeleteSessionByToken(r.Context(), cookie.Value); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to delete session", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	utils.ClearSessionCookie(w)

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}
