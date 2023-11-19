package handler

import (
	"context"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Piszmog/pathwise/components"
	"github.com/Piszmog/pathwise/db/store"
	"github.com/Piszmog/pathwise/types"
	"github.com/Piszmog/pathwise/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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
	jobs, total, err := h.JobApplicationStore.Get(r.Context(), defaultLimitOpts)
	if err != nil {
		h.Logger.Error("failed to get jobs", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	statsOpts, err := h.StatsStore.Get(r.Context())
	if err != nil {
		h.Logger.Error("failed to get stats", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	m := components.Main(
		jobs,
		statsOpts,
		types.PaginationOpts{
			Page:    defaultPage,
			PerPage: defaultPerPage,
			Total:   total,
		},
		types.FilterOpts{},
	)
	m.Render(r.Context(), w)
}

func (h *Handler) GetJobs(w http.ResponseWriter, r *http.Request) {
	page, perPage, err := getPageOpts(r)
	if err != nil {
		h.Logger.Error("failed to get page opts", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	filterOpts := getFilterOpts(r)

	jobs, total, err := h.JobApplicationStore.Filter(r.Context(), store.LimitOpts{Page: page, PerPage: perPage}, filterOpts.Company, filterOpts.Status)
	if err != nil {
		h.Logger.Error("failed to get jobs", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	components.Jobs(
		jobs,
		types.PaginationOpts{Page: page, PerPage: perPage, Total: total},
		filterOpts,
	).Render(r.Context(), w)
}

func (h *Handler) JobDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
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

	details := components.JobDetails(job, timelineEntries)
	details.Render(r.Context(), w)
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
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	if company == "" || title == "" || url == "" {
		h.Logger.Error("missing required form values", "company", company, "title", title, "url", url)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
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
		return
	}

	jobs, total, err := h.JobApplicationStore.Get(r.Context(), defaultLimitOpts)
	if err != nil {
		h.Logger.Error("failed to get jobs", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	stats, err := h.StatsStore.Get(r.Context())
	if err != nil {
		h.Logger.Error("failed to get stats", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	components.AddJob(
		jobs,
		stats,
		types.PaginationOpts{Page: defaultPage, PerPage: defaultPerPage, Total: total},
		types.FilterOpts{},
	).Render(r.Context(), w)
}

func (h *Handler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.Logger.Error("failed to parse id", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	company := r.FormValue("company")
	title := r.FormValue("title")
	url := r.FormValue("url")
	status := r.FormValue("status")
	if company == "" || title == "" || url == "" || status == "" {
		h.Logger.Error("missing required form values", "company", company, "title", title, "url", url, "status", status)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	previousStatus := r.FormValue("previousStatus")
	firstTimelineEntryID := r.FormValue("firstTimelineEntryID")
	firstTimelineEntryType := r.FormValue("firstTimelineEntryType")

	job := types.JobApplication{
		ID:      id,
		Company: company,
		Title:   title,
		URL:     url,
		Status:  types.ToJobApplicationStatus(status),
	}
	updatedAt, err := h.JobApplicationStore.Update(r.Context(), job)
	if err != nil {
		h.Logger.Error("failed to update job", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	job.UpdatedAt = updatedAt

	var stats *types.StatsOpts
	var newTimelineEntry types.NewTimelineEntry
	if previousStatus != status {
		updatedStats, err := h.StatsStore.Get(r.Context())
		if err != nil {
			h.Logger.Error("failed to get stats", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stats = &updatedStats

		latestStatus, err := h.JobApplicationStatusHistoryStore.GetLatestByID(r.Context(), id)
		if err != nil {
			h.Logger.Error("failed to get latest status", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if latestStatus.ID > 0 {
			newTimelineEntry = types.NewTimelineEntry{
				SwapOOB: "beforebegin:#" + newTimelineID(types.ToJobApplicationTimelineType(firstTimelineEntryType), firstTimelineEntryID),
				Entry:   latestStatus,
			}
		}
	}

	components.UpdateJob(job, stats, newTimelineEntry).Render(r.Context(), w)
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
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.Logger.Error("failed to parse id", "error", err)
		w.WriteHeader(http.StatusBadRequest)
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
	components.TimelineEntry(n, false).Render(r.Context(), w)
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

func getFilterOpts(r *http.Request) types.FilterOpts {
	queries := r.URL.Query()
	return types.FilterOpts{
		Company: queries.Get("company"),
		Status:  types.ToJobApplicationStatus(queries.Get("status")),
	}
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	components.Signup().Render(r.Context(), w)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")
	if email == "" || password == "" || confirmPassword == "" {
		h.Logger.Error("missing required form values", "email", email)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if password != confirmPassword {
		h.Logger.Error("passwords do not match", "email", email)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hashedPassword, err := utils.HashPassword([]byte(password))
	if err != nil {
		h.Logger.Error("failed to hash password", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user := types.User{
		Email:    email,
		Password: string(hashedPassword),
	}
	if err := h.UserStore.Insert(r.Context(), user); err != nil {
		h.Logger.Error("failed to insert user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", "/signin")
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func (h *Handler) Signin(w http.ResponseWriter, r *http.Request) {
	components.Signin().Render(r.Context(), w)
}

func (h *Handler) Authenticate(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		h.Logger.Error("missing required form values", "email", email)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.UserStore.GetByEmail(r.Context(), email)
	if err != nil {
		h.Logger.Error("failed to get user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user.ID == 0 {
		h.Logger.Error("user does not exist", "email", email)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err = utils.CheckPasswordHash([]byte(user.Password), []byte(password)); err != nil {
		h.Logger.Error("failed to compare password and hash", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session, err := h.newSession(r.Context(), user.ID)
	if err != nil {
		h.Logger.Error("failed to create session", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
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

func (h *Handler) newSession(ctx context.Context, userId int) (types.Session, error) {
	err := h.SessionsStore.Delete(ctx, userId)
	if err != nil {
		return types.Session{}, err
	}

	token := uuid.New().String()
	session := types.Session{
		UserID:    userId,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := h.SessionsStore.Insert(ctx, session); err != nil {
		return session, err
	}
	return session, nil
}

func (h *Handler) Signout(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("USER-ID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.Logger.Error("failed to parse user id", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.SessionsStore.Delete(r.Context(), userID)
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

	components.Settings(user).Render(r.Context(), w)
}
