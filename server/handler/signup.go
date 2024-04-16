package handler

import (
	"bytes"
	"net/http"
	"regexp"
	"sync"

	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/db/queries"
	"github.com/Piszmog/pathwise/types"
	"github.com/Piszmog/pathwise/utils"
)

var (
	signupHTML []byte
	signupOnce sync.Once
)

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	signupOnce.Do(func() {
		var buf bytes.Buffer
		if err := components.Signup().Render(r.Context(), &buf); err != nil {
			h.Logger.Error("failed to render signup", "error", err)
			return
		}
		signupHTML = buf.Bytes()
	})
	h.htmlStatic(w, http.StatusOK, signupHTML)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")
	if email == "" || password == "" || confirmPassword == "" {
		h.Logger.Debug("missing required form values", "email", email)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Missing email or password", "Please enter your email and password."))
		return
	}
	if password != confirmPassword {
		h.Logger.Debug("passwords do not match", "email", email)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Passwords do not match", "Please enter matching passwords."))
		return
	}

	if !isValidPassword(password) {
		h.Logger.Debug("password does not meet requirements", "email", email)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Password does not meet requirements", "Password must be at least 12 characters long.", "Password must contain at least one uppercase letter.", "Password must contain at least one lowercase letter.", "Password must contain at least one number.", "Password must contain at least one special character (!@#$%^&*)."))
		return
	}

	hashedPassword, err := utils.HashPassword([]byte(password))
	if err != nil {
		h.Logger.Error("failed to hash password", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	user := queries.InsertUserParams{
		Email:    email,
		Password: string(hashedPassword),
	}
	userID, err := h.Database.Queries().InsertUser(r.Context(), user)
	if err != nil {
		h.Logger.Error("failed to insert user", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	if err = h.Database.Queries().InsertNewJobApplicationStat(r.Context(), userID); err != nil {
		h.Logger.Error("failed to insert new job application stat", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	w.Header().Set("HX-Redirect", "/signin")
}

var (
	upperCase   = regexp.MustCompile(`[A-Z]`)
	lowerCase   = regexp.MustCompile(`[a-z]`)
	number      = regexp.MustCompile(`\d`)
	specialChar = regexp.MustCompile(`[!@#$%^&*]`)
)

func isValidPassword(password string) bool {
	if len(password) < 12 {
		return false
	}

	// Check for at least one occurrence of each character class
	hasUpper := upperCase.MatchString(password)
	hasLower := lowerCase.MatchString(password)
	hasNumber := number.MatchString(password)
	hasSpecial := specialChar.MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}
