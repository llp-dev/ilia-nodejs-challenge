package repository_test

import (
	"context"
	"testing"

	"users/internal/repository"
	"users/internal/testhelper"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	pool, cleanup := testhelper.NewPostgresContainer(m)
	testPool = pool
	m.Run()
	cleanup()
}

func TestUserRepository_Create(t *testing.T) {
	testhelper.Truncate(t, testPool)
	repo := repository.NewUserRepository(testPool)
	ctx := context.Background()

	t.Run("creates user", func(t *testing.T) {
		u, err := repo.Create(ctx, "550e8400-e29b-41d4-a716-446655440000", "Alice", "alice@example.com", "hashedpass")
		if err != nil {
			t.Fatalf("Create() error: %v", err)
		}
		if u.ID == "" {
			t.Error("expected non-empty ID")
		}
		if u.Email != "alice@example.com" {
			t.Errorf("Email = %q, want %q", u.Email, "alice@example.com")
		}
	})

	t.Run("returns ErrEmailTaken for duplicate email", func(t *testing.T) {
		_, err := repo.Create(ctx, "550e8400-e29b-41d4-a716-446655440001", "Alice2", "alice@example.com", "hashedpass")
		if err == nil {
			t.Fatal("expected ErrEmailTaken, got nil")
		}
		if err != repository.ErrEmailTaken {
			t.Errorf("error = %v, want ErrEmailTaken", err)
		}
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	testhelper.Truncate(t, testPool)
	repo := repository.NewUserRepository(testPool)
	ctx := context.Background()

	created, _ := repo.Create(ctx, "550e8400-e29b-41d4-a716-446655440000", "Bob", "bob@example.com", "hash")

	t.Run("returns user by id", func(t *testing.T) {
		u, err := repo.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("GetByID() error: %v", err)
		}
		if u.ID != created.ID {
			t.Errorf("ID = %q, want %q", u.ID, created.ID)
		}
	})

	t.Run("returns ErrUserNotFound for unknown id", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
		if err != repository.ErrUserNotFound {
			t.Errorf("error = %v, want ErrUserNotFound", err)
		}
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	testhelper.Truncate(t, testPool)
	repo := repository.NewUserRepository(testPool)
	ctx := context.Background()

	repo.Create(ctx, "550e8400-e29b-41d4-a716-446655440000", "Carol", "carol@example.com", "secrethash")

	t.Run("returns user and hash by email", func(t *testing.T) {
		u, hash, err := repo.GetByEmail(ctx, "carol@example.com")
		if err != nil {
			t.Fatalf("GetByEmail() error: %v", err)
		}
		if u.Email != "carol@example.com" {
			t.Errorf("Email = %q, want carol@example.com", u.Email)
		}
		if hash != "secrethash" {
			t.Errorf("hash = %q, want secrethash", hash)
		}
	})

	t.Run("returns ErrUserNotFound for unknown email", func(t *testing.T) {
		_, _, err := repo.GetByEmail(ctx, "nobody@example.com")
		if err != repository.ErrUserNotFound {
			t.Errorf("error = %v, want ErrUserNotFound", err)
		}
	})
}

func TestUserRepository_UpdateProfile(t *testing.T) {
	testhelper.Truncate(t, testPool)
	repo := repository.NewUserRepository(testPool)
	ctx := context.Background()

	u, _ := repo.Create(ctx, "550e8400-e29b-41d4-a716-446655440000", "Dave", "dave@example.com", "hash")

	t.Run("updates name and email", func(t *testing.T) {
		updated, err := repo.UpdateProfile(ctx, u.ID, "David", "david@example.com")
		if err != nil {
			t.Fatalf("UpdateProfile() error: %v", err)
		}
		if updated.Name != "David" {
			t.Errorf("Name = %q, want David", updated.Name)
		}
		if updated.Email != "david@example.com" {
			t.Errorf("Email = %q, want david@example.com", updated.Email)
		}
	})
}
