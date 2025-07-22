package handler

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/db/queries"
	"github.com/Piszmog/pathwise/types"
	"github.com/Piszmog/pathwise/utils"
	"github.com/google/uuid"
)

var (
	signinHTML []byte
	signinOnce sync.Once
)

func (h *Handler) Signin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	signinOnce.Do(func() {
		var buf bytes.Buffer
		if err := components.Signin().Render(ctx, &buf); err != nil {
			h.Logger.ErrorContext(ctx, "failed to render signin", "error", err)
			return
		}
		signinHTML = buf.Bytes()
	})
	h.htmlStatic(ctx, w, http.StatusOK, signinHTML)
}

func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		h.Logger.DebugContext(r.Context(), "missing required form values", "email", email)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Missing email or password", "Please enter your email and password."))
		return
	}
	user, err := h.Database.Queries().GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.Logger.DebugContext(r.Context(), "user not found", "email", email)
			h.html(r.Context(), w, http.StatusUnauthorized, components.Alert(types.AlertTypeError, "Incorrect email or password", "Double check your email and password and try again."))
		} else {
			h.Logger.ErrorContext(r.Context(), "failed to get user", "error", err)
			h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeWarning, "Something went wrong", "Try again later."))
		}
		return
	}
	if err = utils.CheckPasswordHash([]byte(user.Password), []byte(password)); err != nil {
		h.Logger.DebugContext(r.Context(), "failed to compare password and hash", "error", err)
		h.html(r.Context(), w, http.StatusForbidden, components.Alert(types.AlertTypeError, "Incorrect email or password", "Double check your email and password and try again."))
		return
	}

	var cookieValue string
	cookie, err := r.Cookie("session")
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			h.Logger.ErrorContext(r.Context(), "failed to get session cookie", "error", err)
		}
	} else {
		cookieValue = cookie.Value
	}

	token, expiresAt, err := h.newSession(r.Context(), user.ID, r.UserAgent(), cookieValue, getClientIP(r))
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to create session", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeWarning, "Something went wrong", "Try again later."))
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	err = h.Database.Queries().DeleteOldUserSessions(r.Context(), user.ID)
	if err != nil {
		h.Logger.WarnContext(r.Context(), "failed to delete old user sessions", "userID", user.ID, "error", err)
	}

	w.Header().Set("HX-Redirect", "/")
}

func (h *Handler) newSession(ctx context.Context, userID int64, userAgent string, currentToken string, ipAddress string) (string, time.Time, error) {
	if currentToken != "" {
		if err := h.Database.Queries().DeleteSessionByToken(ctx, currentToken); err != nil {
			return "", time.Time{}, err
		}
	}

	token := uuid.New().String()
	session := queries.InsertSessionParams{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserAgent: userAgent,
		IpAddress: ipAddress,
	}
	if err := h.Database.Queries().InsertSession(ctx, session); err != nil {
		return "", time.Time{}, err
	}
	return session.Token, session.ExpiresAt, nil
}
