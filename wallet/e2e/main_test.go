package e2e_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"wallet/internal/server"
	"wallet/internal/testhelper"
)

var testServer *httptest.Server
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	pool, cleanup := testhelper.NewPostgresContainer(m)
	testPool = pool

	s := server.New(pool)
	testServer = httptest.NewServer(s.Handler())

	code := m.Run()

	testServer.Close()
	cleanup()
	os.Exit(code)
}
