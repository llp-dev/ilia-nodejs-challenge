package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"users/internal/server"
	"users/internal/testhelper"
)

const (
	testSecret         = "test-secret"
	testInternalSecret = "test-internal-secret"
)

var testServer *httptest.Server
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	pool, cleanup := testhelper.NewPostgresContainer(m)
	testPool = pool

	s := server.New(pool, testSecret, testInternalSecret)
	testServer = httptest.NewServer(s.Handler())

	code := m.Run()

	testServer.Close()
	cleanup()
	os.Exit(code)
}

func req(method, url string, body interface{ Read([]byte) (int, error) }, token string) *http.Request {
	var r *http.Request
	if body != nil {
		r, _ = http.NewRequest(method, url, body)
	} else {
		r, _ = http.NewRequest(method, url, nil)
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	r.Header.Set("Content-Type", "application/json")
	return r
}
