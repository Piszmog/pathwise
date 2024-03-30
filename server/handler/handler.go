package handler

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/types"
	"github.com/Piszmog/pathwise/utils"
	"github.com/google/uuid"
)

const (
	defaultPage    = 0
	defaultPerPage = 10
)

var defaultLimitOpts = store.LimitOpts{Page: defaultPage, PerPage: defaultPerPage}

type Handler struct {
	Logger                           *slog.Logger
	JobApplicationStore              *store.JobApplicationStore
	JobApplicationNoteStore          *store.JobApplicationNoteStore
	JobApplicationStatusHistoryStore *store.JobApplicationStatusHistoryStore
	StatsStore                       *store.StatsStore
	UserStore                        *store.UserStore
	SessionsStore                    *store.SessionStore
}

func (h *Handler) Main(w http.ResponseWriter, r *http.Request) {
	_ = components.Main().Render(r.Context(), w)
}

func (h *Handler) GetJobs(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	page, perPage, err := getPageOpts(r)
	if err != nil {
		h.Logger.Error("failed to get page opts", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	filterOpts := getFilterOpts(r)

	jobs, total, err := h.JobApplicationStore.Filter(r.Context(), store.LimitOpts{Page: page, PerPage: perPage}, userID, filterOpts.Company, filterOpts.Status)
	if err != nil {
		h.Logger.Error("failed to get jobs", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = components.Jobs(
		jobs,
		types.PaginationOpts{Page: page, PerPage: perPage, Total: total},
		filterOpts,
	).Render(r.Context(), w)
}

func getFilterOpts(r *http.Request) types.FilterOpts {
	queries := r.URL.Query()
	return types.FilterOpts{
		Company: queries.Get("company"),
		Status:  types.ToJobApplicationStatus(queries.Get("status")),
	}
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	statsOpts, err := h.StatsStore.Get(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to get stats", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = components.Stats(statsOpts, false, "").Render(r.Context(), w)
}

func (h *Handler) JobDetails(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.Logger.Error("failed to parse id", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	job, err := h.JobApplicationStore.GetByID(r.Context(), id)
	if err != nil {
		h.Logger.Error("failed to get job", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	timelineEntries, err := h.getTimelineEntries(r.Context(), id)
	if err != nil {
		h.Logger.Error("failed to get timeline entries", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = components.JobDetails(job, timelineEntries).Render(r.Context(), w)
}

func (h *Handler) getTimelineEntries(ctx context.Context, id int) ([]types.JobApplicationTimelineEntry, error) {
	notes, err := h.JobApplicationNoteStore.GetAllByID(ctx, id)
	if err != nil {
		return nil, err
	}
	histories, err := h.JobApplicationStatusHistoryStore.GetAllByID(ctx, id)
	if err != nil {
		return nil, err
	}
	timelineEntries := make([]types.JobApplicationTimelineEntry, len(notes)+len(histories))
	for i, note := range notes {
		timelineEntries[i] = note
	}
	for i, history := range histories {
		timelineEntries[i+len(notes)] = history
	}
	sort.Slice(timelineEntries, func(i, j int) bool {
		return timelineEntries[i].Created().After(timelineEntries[j].Created())
	})
	return timelineEntries, nil
}

func (h *Handler) AddJob(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	if company == "" || title == "" || url == "" {
		h.Logger.Warn("missing required form values", "company", company, "title", title, "url", url)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Missing company, title, or url", "Please enter a company, title, and url.").Render(r.Context(), w)
		return
	}

	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	job := types.JobApplication{
		Company: company,
		Title:   title,
		URL:     url,
		UserID:  userID,
	}
	if err = h.JobApplicationStore.Insert(r.Context(), job); err != nil {
		h.Logger.Error("failed to insert job", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	jobs, total, err := h.JobApplicationStore.Get(r.Context(), userID, defaultLimitOpts)
	if err != nil {
		h.Logger.Error("failed to get jobs", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	stats, err := h.StatsStore.Get(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to get stats", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	_ = components.AddJob(
		jobs,
		stats,
		types.PaginationOpts{Page: defaultPage, PerPage: defaultPerPage, Total: total},
		types.FilterOpts{},
	).Render(r.Context(), w)
}

func (h *Handler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.Logger.Error("failed to parse id", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	job, err := h.JobApplicationStore.GetByIDAndUserID(r.Context(), id, userID)
	if err != nil {
		h.Logger.Error("failed to get job", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	if job.UserID != userID {
		h.Logger.Warn("user does not own job", "userID", userID, "jobUserID", job.UserID)
		w.WriteHeader(http.StatusForbidden)
		_ = components.Alert(types.AlertTypeError, "You do not have permission to update this job", "Try again later.").Render(r.Context(), w)
		return
	}

	if err = r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	status := r.FormValue("status")
	if company == "" || title == "" || url == "" || status == "" {
		h.Logger.Warn("missing required form values", "company", company, "title", title, "url", url, "status", status)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Missing company, title, url, or status", "Please enter a company, title, url, and status.").Render(r.Context(), w)
		return
	}
	previousStatus := r.FormValue("previousStatus")
	firstTimelineEntryID := r.FormValue("firstTimelineEntryID")
	firstTimelineEntryType := r.FormValue("firstTimelineEntryType")

	job.Company = company
	job.Title = title
	job.URL = url
	job.Status = types.ToJobApplicationStatus(status)
	updatedAt, err := h.JobApplicationStore.Update(r.Context(), job)
	if err != nil {
		h.Logger.Error("failed to update job", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}
	job.UpdatedAt = updatedAt

	var stats *types.StatsOpts
	var newTimelineEntry types.NewTimelineEntry
	if previousStatus != status {
		updatedStats, err := h.StatsStore.Get(r.Context(), userID)
		if err != nil {
			h.Logger.Error("failed to get stats", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
			return
		}
		stats = &updatedStats

		latestStatus, err := h.JobApplicationStatusHistoryStore.GetLatestByID(r.Context(), id)
		if err != nil {
			h.Logger.Error("failed to get latest status", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
			return
		}
		if latestStatus.ID > 0 {
			newTimelineEntry = types.NewTimelineEntry{
				SwapOOB: "beforebegin:#" + newTimelineID(types.ToJobApplicationTimelineType(firstTimelineEntryType), firstTimelineEntryID),
				Entry:   latestStatus,
			}
		}
	}

	_ = components.UpdateJob(job, stats, newTimelineEntry).Render(r.Context(), w)
}

func newTimelineID(entryType types.JobApplicationTimelineType, entryID string) string {
	switch entryType {
	case types.JobApplicationTimelineTypeStatus:
		return utils.TimelineStatusRowStringID(entryID)
	case types.JobApplicationTimelineTypeNote:
		return utils.TimelineNoteRowStringID(entryID)
	default:
		return "unknown"
	}
}

func (h *Handler) AddNote(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.Logger.Error("failed to parse id", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	job, err := h.JobApplicationStore.GetByIDAndUserID(r.Context(), id, userID)
	if err != nil {
		h.Logger.Error("failed to get job", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if job.UserID != userID {
		h.Logger.Warn("user does not own job", "userID", userID, "jobUserID", job.UserID)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err = r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	note := r.FormValue("note")
	if note == "" {
		h.Logger.Error("missing required form values", "note", note)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jobNote := types.JobApplicationNote{
		JobApplicationID: id,
		Note:             note,
	}
	n, err := h.JobApplicationNoteStore.Insert(r.Context(), jobNote)
	if err != nil {
		h.Logger.Error("failed to insert note", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = components.TimelineEntry(n, false).Render(r.Context(), w)
}

func getPageOpts(r *http.Request) (int, int, error) {
	queries := r.URL.Query()
	pageQuery := queries.Get("page")
	perPageQuery := queries.Get("per_page")
	page := defaultPage
	var err error
	if pageQuery != "" {
		page, err = strconv.Atoi(pageQuery)
		if err != nil {
			return 0, 0, err
		}
	}
	perPage := defaultPerPage
	if perPageQuery != "" {
		perPage, err = strconv.Atoi(perPageQuery)
		if err != nil {
			return 0, 0, err
		}
	}
	return page, perPage, nil
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	_ = components.Signup().Render(r.Context(), w)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")
	if email == "" || password == "" || confirmPassword == "" {
		h.Logger.Debug("missing required form values", "email", email)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Missing email or password", "Please enter your email and password.").Render(r.Context(), w)
		return
	}
	if password != confirmPassword {
		h.Logger.Debug("passwords do not match", "email", email)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Passwords do not match", "Please enter matching passwords.").Render(r.Context(), w)
		return
	}

	if !isValidPassword(password) {
		h.Logger.Debug("password does not meet requirements", "email", email)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Password does not meet requirements", "Password must be at least 12 characters long.", "Password must contain at least one uppercase letter.", "Password must contain at least one lowercase letter.", "Password must contain at least one number.", "Password must contain at least one special character (!@#$%^&*).").Render(r.Context(), w)
		return
	}

	hashedPassword, err := utils.HashPassword([]byte(password))
	if err != nil {
		h.Logger.Error("failed to hash password", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}
	user := types.User{
		Email:    email,
		Password: string(hashedPassword),
	}
	if err := h.UserStore.Insert(r.Context(), user); err != nil {
		h.Logger.Error("failed to insert user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
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

func (h *Handler) Signin(w http.ResponseWriter, r *http.Request) {
	_ = components.Signin().Render(r.Context(), w)
}

func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		h.Logger.Debug("missing required form values", "email", email)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Missing emil or password", "Please enter your email and password.").Render(r.Context(), w)
		return
	}
	user, err := h.UserStore.GetByEmail(r.Context(), email)
	if err != nil {
		if err == sql.ErrNoRows {
			h.Logger.Debug("user not found", "email", email)
			w.WriteHeader(http.StatusUnauthorized)
			_ = components.Alert(types.AlertTypeError, "Incorrect email or password", "Double check your email and password and try again.").Render(r.Context(), w)
		} else {
			h.Logger.Error("failed to get user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = components.Alert(types.AlertTypeWarning, "Something went wrong", "Try again later.").Render(r.Context(), w)
		}
		return
	}
	if err = utils.CheckPasswordHash([]byte(user.Password), []byte(password)); err != nil {
		h.Logger.Debug("failed to compare password and hash", "error", err)
		w.WriteHeader(http.StatusForbidden)
		_ = components.Alert(types.AlertTypeError, "Incorrect email or password", "Double check your email and password and try again.").Render(r.Context(), w)
		return
	}

	var cookieValue string
	cookie, err := r.Cookie("session")
	if err != nil {
		if err != http.ErrNoCookie {
			h.Logger.Error("failed to get session cookie", "error", err)
		}
	} else {
		cookieValue = cookie.Value
	}

	session, err := h.newSession(r.Context(), user.ID, r.UserAgent(), cookieValue)
	if err != nil {
		h.Logger.Error("failed to create session", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeWarning, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.Token,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("HX-Redirect", "/")
}

func (h *Handler) newSession(ctx context.Context, userId int, userAgent string, currentToken string) (types.Session, error) {
	if currentToken != "" {
		err := h.SessionsStore.DeleteByToken(ctx, currentToken)
		if err != nil {
			return types.Session{}, err
		}
	}

	token := uuid.New().String()
	session := types.Session{
		UserID:    userId,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserAgent: userAgent,
	}
	if err := h.SessionsStore.Insert(ctx, session); err != nil {
		return session, err
	}
	return session, nil
}

func (h *Handler) Signout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		h.Logger.Error("failed to get session cookie", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.SessionsStore.DeleteByToken(r.Context(), cookie.Value)
	if err != nil {
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

func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := h.UserStore.GetByID(r.Context(), int64(userID))
	if err != nil {
		h.Logger.Error("failed to get user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = components.Settings(user).Render(r.Context(), w)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	user, err := h.UserStore.GetByID(r.Context(), int64(userID))
	if err != nil {
		h.Logger.Error("failed to get user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	if err = r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	currentPassword := r.FormValue("currentPassword")
	newPassword := r.FormValue("newPassword")
	confirmPassword := r.FormValue("confirmPassword")

	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		h.Logger.Debug("missing required form values")
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Missing current password, new password, or confirm new password", "Please enter your current password, new password, and confirm new password.").Render(r.Context(), w)
		return
	}

	if newPassword != confirmPassword {
		h.Logger.Debug("passwords do not match", "newPassword", newPassword, "confirmNewPassword", confirmPassword)
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Passwords do not match", "Please enter matching passwords.").Render(r.Context(), w)
		return
	}

	if currentPassword == newPassword {
		h.Logger.Debug("current password and new password are the same")
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "New password cannot be the same as current password", "Please enter a new password.").Render(r.Context(), w)
		return
	}

	if !isValidPassword(newPassword) {
		h.Logger.Debug("password does not meet requirements")
		w.WriteHeader(http.StatusBadRequest)
		_ = components.Alert(types.AlertTypeError, "Password does not meet requirements", "Password must be at least 12 characters long.", "Password must contain at least one uppercase letter.", "Password must contain at least one lowercase letter.", "Password must contain at least one number.", "Password must contain at least one special character (!@#$%^&*).").Render(r.Context(), w)
		return
	}

	if err = utils.CheckPasswordHash([]byte(user.Password), []byte(currentPassword)); err != nil {
		h.Logger.Debug("failed to compare password and hash", "error", err)
		w.WriteHeader(http.StatusForbidden)
		_ = components.Alert(types.AlertTypeError, "Incorrect password", "Double check your password and try again.").Render(r.Context(), w)
		return
	}

	hashedPassword, err := utils.HashPassword([]byte(newPassword))
	if err != nil {
		h.Logger.Error("failed to hash password", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	if err = h.UserStore.UpdatePassword(r.Context(), int64(userID), string(hashedPassword)); err != nil {
		h.Logger.Error("failed to update password", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	err = h.SessionsStore.DeleteByUserID(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to delete sessions", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
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
	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	err = h.SessionsStore.DeleteByUserID(r.Context(), userID)
	if err != nil {
		h.Logger.Error("failed to delete sessions", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
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
	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
		return
	}

	err = h.UserStore.Delete(r.Context(), int64(userID))
	if err != nil {
		h.Logger.Error("failed to delete user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = components.Alert(types.AlertTypeError, "Something went wrong", "Try again later.").Render(r.Context(), w)
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
