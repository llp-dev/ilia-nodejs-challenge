package e2e_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"wallet/internal/server"
	"wallet/internal/testhelper"
)

const (
	testSecret    = "test-secret"
	testUserID    = "550e8400-e29b-41d4-a716-446655440000"
	testUserEmail = "user@example.com"
)

var testServer *httptest.Server
var testPool *pgxpool.Pool
var testToken string

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   testUserID,
		"email": testUserEmail,
	}).SignedString([]byte(testSecret))
	testToken = tok

	pool, cleanup := testhelper.NewPostgresContainer(m)
	testPool = pool

	// Fake users service: GET /users/:id returns {id, email}
	fakeUsers := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/users/")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": id, "email": testUserEmail})
	}))
	defer fakeUsers.Close()

	s := server.New(pool, testSecret, testSecret, fakeUsers.URL)
	testServer = httptest.NewServer(s.Handler())

	code := m.Run()

	testServer.Close()
	cleanup()
	os.Exit(code)
}

func authReq(method, url string, body interface{ Read([]byte) (int, error) }) *http.Request {
	var req *http.Request
	if body != nil {
		req, _ = http.NewRequest(method, url, body)
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}
	req.Header.Set("Authorization", "Bearer "+testToken)
	req.Header.Set("Content-Type", "application/json")
	return req
}
