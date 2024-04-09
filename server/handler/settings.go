package handler

import (
	"net/http"
	"time"

	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/db/queries"
	"github.com/Piszmog/pathwise/types"
	"github.com/Piszmog/pathwise/utils"
)

func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	user, err := h.Database.Queries().GetUserByID(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to get user", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	h.html(r.Context(), w, http.StatusOK, components.Settings(user.Email))
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	user, err := h.Database.Queries().GetUserByID(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to get user", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	currentPassword := r.FormValue("currentPassword")
	newPassword := r.FormValue("newPassword")
	confirmPassword := r.FormValue("confirmPassword")

	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		h.Logger.Debug("missing required form values")
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Missing current password, new password, or confirm new password", "Please enter your current password, new password, and confirm new password."))
		return
	}

	if newPassword != confirmPassword {
		h.Logger.Debug("passwords do not match", "newPassword", newPassword, "confirmNewPassword", confirmPassword)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Passwords do not match", "Please enter matching passwords."))
		return
	}

	if currentPassword == newPassword {
		h.Logger.Debug("current password and new password are the same")
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "New password cannot be the same as current password", "Please enter a new password."))
		return
	}

	if !isValidPassword(newPassword) {
		h.Logger.Debug("password does not meet requirements")
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Password does not meet requirements", "Password must be at least 12 characters long.", "Password must contain at least one uppercase letter.", "Password must contain at least one lowercase letter.", "Password must contain at least one number.", "Password must contain at least one special character (!@#$%^&*)."))
		return
	}

	if err = utils.CheckPasswordHash([]byte(user.Password), []byte(currentPassword)); err != nil {
		h.Logger.Debug("failed to compare password and hash", "error", err)
		h.html(r.Context(), w, http.StatusForbidden, components.Alert(types.AlertTypeError, "Incorrect password", "Double check your password and try again."))
		return
	}

	hashedPassword, err := utils.HashPassword([]byte(newPassword))
	if err != nil {
		h.Logger.Error("failed to hash password", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = h.Database.Queries().UpdateUserPassword(r.Context(), queries.UpdateUserPasswordParams{ID: userID, Password: string(hashedPassword)}); err != nil {
		h.Logger.Error("failed to update password", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = h.Database.Queries().DeleteSessionByUserID(r.Context(), userID); err != nil {
		h.Logger.Error("failed to delete sessions", "error", err)
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
		h.Logger.Error("failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = h.Database.Queries().DeleteSessionByUserID(r.Context(), userID); err != nil {
		h.Logger.Error("failed to delete sessions", "error", err)
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
		h.Logger.Error("failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if err = h.Database.Queries().DeleteUserByID(r.Context(), userID); err != nil {
		h.Logger.Error("failed to delete user", "error", err)
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
