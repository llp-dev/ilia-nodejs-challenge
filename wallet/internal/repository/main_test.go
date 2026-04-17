package repository_test

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"wallet/internal/testhelper"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	pool, cleanup := testhelper.NewPostgresContainer(m)
	testPool = pool

	code := m.Run()

	cleanup()
	os.Exit(code)
}
