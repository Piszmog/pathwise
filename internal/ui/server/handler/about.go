package handler

import (
	"bytes"
	"net/http"
	"sync"

	"github.com/Piszmog/pathwise/internal/ui/components"
)

var (
	aboutHTML []byte
	aboutOnce sync.Once
)

func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	aboutOnce.Do(func() {
		var buf bytes.Buffer
		if err := components.About().Render(ctx, &buf); err != nil {
			h.Logger.ErrorContext(ctx, "failed to render about", "error", err)
			return
		}
		aboutHTML = buf.Bytes()
	})
	h.htmlStatic(ctx, w, http.StatusOK, aboutHTML)
}
