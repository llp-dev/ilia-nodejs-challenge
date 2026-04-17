package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestServer_Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{
			name:           "GET /health returns 200",
			method:         http.MethodGet,
			path:           "/health",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "unknown route returns 404",
			method:         http.MethodGet,
			path:           "/unknown",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "POST /health returns 404",
			method:         http.MethodPost,
			path:           "/health",
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(nil)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			s.router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.wantStatusCode)
			}
		})
	}
}
