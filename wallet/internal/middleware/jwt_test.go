package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"wallet/internal/middleware"
)

func TestJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "ILIACHALLENGE"

	validToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1"}).SignedString([]byte(secret))

	tests := []struct {
		name   string
		header string
		want   int
	}{
		{"no header", "", http.StatusUnauthorized},
		{"bad format", "Token abc", http.StatusUnauthorized},
		{"wrong secret", "Bearer " + func() string { t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{}).SignedString([]byte("wrong")); return t }(), http.StatusUnauthorized},
		{"valid token", "Bearer " + validToken, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/", middleware.JWT(secret), func(c *gin.Context) { c.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			r.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("got %d, want %d", w.Code, tt.want)
			}
		})
	}
}
