package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"wallet/internal/server"
	"wallet/internal/testhelper"
)

const testSecret = "test-secret"

var testServer *httptest.Server
var testPool *pgxpool.Pool
var testToken string

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{}).SignedString([]byte(testSecret))
	testToken = tok

	pool, cleanup := testhelper.NewPostgresContainer(m)
	testPool = pool

	s := server.New(pool, testSecret)
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
