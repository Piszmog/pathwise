package handler_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Piszmog/pathwise/server/handler"
	"github.com/Piszmog/pathwise/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Health(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:           "successful GET request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]string{
				"status":  "ok",
				"version": version.Value,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			h := &handler.Handler{
				Logger:   logger,
				Database: nil,
			}

			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			h.Health(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}
