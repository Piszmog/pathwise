package handler

import (
	"net/http"

	"github.com/Piszmog/pathwise/ui/components"
)

func (h *Handler) Archives(w http.ResponseWriter, r *http.Request) {
	h.html(r.Context(), w, http.StatusOK, components.Archives())
}
