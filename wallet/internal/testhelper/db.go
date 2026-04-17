package testhelper

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"wallet/internal/db"
)

// NewPostgresContainer starts a single PostgreSQL container intended for use in TestMain.
// Returns the pool and a cleanup function that terminates the container and closes the pool.
func NewPostgresContainer(m *testing.M) (*pgxpool.Pool, func()) {
	// Ryuk requires mounting /var/run/docker.sock which is unavailable with rootless Podman.
	// Cleanup is handled by the returned func, so the reaper is not needed.
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("wallet_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		panic(fmt.Sprintf("start postgres container: %v", err))
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		panic(fmt.Sprintf("get connection string: %v", err))
	}

	err = db.Migrate(dsn)
	if err != nil {
		container.Terminate(ctx)
		panic(fmt.Sprintf("run migrations: %v", err))
	}

	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		container.Terminate(ctx)
		panic(fmt.Sprintf("connect to db: %v", err))
	}

	cleanup := func() {
		pool.Close()
		container.Terminate(ctx)
	}

	return pool, cleanup
}

// Truncate clears all tables between tests to ensure isolation.
func Truncate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), "TRUNCATE TABLE transactions, wallets CASCADE")
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}
