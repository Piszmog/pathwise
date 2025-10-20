package middleware

import (
	"context"
	"net/http"

	contextkey "github.com/Piszmog/pathwise/internal/context_key"
	"github.com/google/uuid"
)

type CorrelationIDMiddleware struct{}

func (m *CorrelationIDMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Correlation-ID")
		if id == "" {
			id = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), contextkey.KeyCorrelationID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
