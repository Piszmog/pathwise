package handler

import (
	"net/http"
	"strconv"

	"github.com/Piszmog/pathwise/ui/components"
	"github.com/Piszmog/pathwise/ui/db/queries"
	"github.com/Piszmog/pathwise/ui/types"
)

func (h *Handler) AddNote(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse user id", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse id", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	job, err := h.Database.Queries().GetJobApplicationByIDAndUserID(r.Context(), queries.GetJobApplicationByIDAndUserIDParams{ID: int64(id), UserID: userID})
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to get job", "error", err)
		h.html(r.Context(), w, http.StatusInternalServerError, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}

	if job.UserID != userID {
		h.Logger.WarnContext(r.Context(), "user does not own job", "userID", userID, "jobUserID", job.UserID)
		h.html(r.Context(), w, http.StatusForbidden, components.Alert(types.AlertTypeError, "You do not have permission to add a note", "You can only add notes to your own job applications."))
		return
	}

	if err = r.ParseForm(); err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to parse form", "error", err)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Something went wrong", "Try again later."))
		return
	}
	note := r.FormValue("note")
	if note == "" {
		h.Logger.ErrorContext(r.Context(), "missing required form values", "note", note)
		h.html(r.Context(), w, http.StatusBadRequest, components.Alert(types.AlertTypeError, "Missing note", "Please enter a note."))
		return
	}

	jobNote := queries.InsertJobApplicationNoteParams{
		JobApplicationID: int64(id),
		Note:             note,
	}
	n, err := h.Database.Queries().InsertJobApplicationNote(r.Context(), jobNote)
	if err != nil {
		h.Logger.ErrorContext(r.Context(), "failed to insert note", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	noteEntry := types.JobApplicationNote{
		ID:               n.ID,
		JobApplicationID: n.JobApplicationID,
		Note:             n.Note,
		CreatedAt:        n.CreatedAt,
	}

	h.html(r.Context(), w, http.StatusOK, components.TimelineEntry(noteEntry, false))
}
