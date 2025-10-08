package handler

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/Piszmog/pathwise/internal/ui/components"
	"github.com/Piszmog/pathwise/internal/ui/types"
	"github.com/Piszmog/pathwise/internal/ui/utils"
	"github.com/google/uuid"
)

func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	user, err := h.Database.Queries().GetUserByID(r.Context(), userID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get user", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	mcpAPIKeyCreatedAt, err := h.Database.Queries().GetMcpAPIKeyByUserID(r.Context(), userID)
	hasMcpAPIKey := false
	mcpKeyCreatedAt := ""
	if err == nil {
		hasMcpAPIKey = true
		mcpKeyCreatedAt = mcpAPIKeyCreatedAt.Format("January 2, 2006")
	} else if !errors.Is(err, sql.ErrNoRows) {
		h.Logger.ErrorContext(r.Context(), "failed to check MCP API key", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.Settings(user.Email, hasMcpAPIKey, mcpKeyCreatedAt))
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	user, err := h.Database.Queries().GetUserByID(r.Context(), userID)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get user", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = r.ParseForm(); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse form", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	currentPassword := r.FormValue("currentPassword")
	newPassword := r.FormValue("newPassword")
	confirmPassword := r.FormValue("confirmPassword")

	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		h.Logger.DebugContext(r.Context(), "missing required form values")
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Missing current password, new password, or confirm new password", "Please enter your current password, new password, and confirm new password."))
		return
	}

	if newPassword != confirmPassword {
		h.Logger.DebugContext(r.Context(), "passwords do not match", "newPassword", newPassword, "confirmNewPassword", confirmPassword)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Passwords do not match", "Please enter matching passwords."))
		return
	}

	if currentPassword == newPassword {
		h.Logger.DebugContext(r.Context(), "current password and new password are the same")
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "New password cannot be the same as current password", "Please enter a new password."))
		return
	}

	if !isValidPassword(newPassword) {
		h.Logger.DebugContext(r.Context(), "password does not meet requirements")
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Password does not meet requirements", "Password must be at least 12 characters long.", "Password must contain at least one uppercase letter.", "Password must contain at least one lowercase letter.", "Password must contain at least one number.", "Password must contain at least one special character (!@#$%^&*)."))
		return
	}

	if err = utils.CheckPasswordHash([]byte(user.Password), []byte(currentPassword)); err != nil {
		h.Logger.DebugContext(r.Context(), "failed to compare password and hash", "error", err)
		h.html(r.Context(), w, http.StatusForbidden, components.Alert(types.AlertTypeError, "Incorrect password", "Double check your password and try again."))
		return
	}

	hashedPassword, err := utils.HashPassword([]byte(newPassword), 14)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to hash password", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = h.Database.Queries().UpdateUserPassword(r.Context(), queries.UpdateUserPasswordParams{ID: userID, Password: string(hashedPassword)}); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to update password", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = h.Database.Queries().DeleteSessionByUserID(r.Context(), userID); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to delete sessions", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("HX-Redirect", "/signin")
}

func (h *Handler) LogoutSessions(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = h.Database.Queries().DeleteSessionByUserID(r.Context(), userID); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to delete sessions", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("HX-Redirect", "/signin")
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = h.Database.Queries().DeleteUserByID(r.Context(), userID); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to delete user", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("HX-Redirect", "/signin")
}

func (h *Handler) CreateMcpAuth(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	_, err = h.Database.Queries().GetMcpAPIKeyByUserID(r.Context(), userID)
	if err == nil {
		h.Logger.DebugContext(r.Context(), "user already has MCP API key")
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "API key already exists", "You already have an MCP API key. Delete it first to create a new one."))
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		h.Logger.ErrorContext(r.Context(), "failed to check existing MCP API key", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	plainAPIKey := uuid.New().String()
	keyHash := fmt.Sprintf("%x", sha256.Sum256([]byte(plainAPIKey)))

	result, err := h.Database.Queries().InsertMcpAPIKey(r.Context(), queries.InsertMcpAPIKeyParams{
		UserID:  userID,
		KeyHash: keyHash,
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to create MCP API key", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	createdAt := result.CreatedAt.Format("January 2, 2006")

	h.html(r.Context(), w, http.StatusOK, components.McpAuthModalWithSectionUpdate(plainAPIKey, createdAt))
}

func (h *Handler) RegenerateMcpAuth(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = h.Database.Queries().DeleteMcpAPIKeyByUserID(r.Context(), userID); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to delete existing MCP API key", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	plainAPIKey := uuid.New().String()
	keyHash := fmt.Sprintf("%x", sha256.Sum256([]byte(plainAPIKey)))

	result, err := h.Database.Queries().InsertMcpAPIKey(r.Context(), queries.InsertMcpAPIKeyParams{
		UserID:  userID,
		KeyHash: keyHash,
	})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to create new MCP API key", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	createdAt := result.CreatedAt.Format("January 2, 2006")

	h.html(r.Context(), w, http.StatusOK, components.McpAuthModalWithSectionUpdate(plainAPIKey, createdAt))
}

func (h *Handler) DeleteMcpAuth(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = h.Database.Queries().DeleteMcpAPIKeyByUserID(r.Context(), userID); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to delete MCP API key", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.McpAuthSection(false, ""))
}
