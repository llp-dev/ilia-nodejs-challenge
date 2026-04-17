package repository_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"wallet/internal/repository"
	"wallet/internal/testhelper"
)

func TestWalletRepository_Create(t *testing.T) {
	testhelper.Truncate(t, testPool)
	repo := repository.NewWalletRepository(testPool)
	ctx := context.Background()

	wallet, err := repo.Create(ctx, "550e8400-e29b-41d4-a716-446655440000", "my wallet")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if wallet.ID == "" {
		t.Error("expected non-empty ID")
	}
	if wallet.UserID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("UserID = %q, want %q", wallet.UserID, "550e8400-e29b-41d4-a716-446655440000")
	}
	if wallet.Description != "my wallet" {
		t.Errorf("Description = %q, want %q", wallet.Description, "my wallet")
	}
	if !wallet.Balance.Equal(decimal.Zero) {
		t.Errorf("Balance = %v, want 0", wallet.Balance)
	}
}

func TestWalletRepository_GetByID(t *testing.T) {
	testhelper.Truncate(t, testPool)
	repo := repository.NewWalletRepository(testPool)
	ctx := context.Background()

	created, err := repo.Create(ctx, "550e8400-e29b-41d4-a716-446655440000", "test wallet")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	t.Run("existing wallet", func(t *testing.T) {
		found, err := repo.GetByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if found.ID != created.ID {
			t.Errorf("ID = %q, want %q", found.ID, created.ID)
		}
	})

	t.Run("non-existent wallet", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
		if err == nil {
			t.Fatal("expected error for non-existent wallet, got nil")
		}
	})
}

func TestWalletRepository_List(t *testing.T) {
	testhelper.Truncate(t, testPool)
	repo := repository.NewWalletRepository(testPool)
	ctx := context.Background()

	_, err := repo.Create(ctx, "550e8400-e29b-41d4-a716-446655440000", "wallet one")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, err = repo.Create(ctx, "550e8400-e29b-41d4-a716-446655440001", "wallet two")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	wallets, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(wallets) != 2 {
		t.Errorf("List() returned %d wallets, want 2", len(wallets))
	}
}

func TestWalletRepository_UpdateDescription(t *testing.T) {
	testhelper.Truncate(t, testPool)
	repo := repository.NewWalletRepository(testPool)
	ctx := context.Background()

	created, err := repo.Create(ctx, "550e8400-e29b-41d4-a716-446655440000", "original")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	t.Run("updates description", func(t *testing.T) {
		updated, err := repo.UpdateDescription(ctx, created.ID, "updated")
		if err != nil {
			t.Fatalf("UpdateDescription() error = %v", err)
		}
		if updated.Description != "updated" {
			t.Errorf("Description = %q, want %q", updated.Description, "updated")
		}
		if !updated.UpdatedAt.After(created.UpdatedAt) {
			t.Error("expected UpdatedAt to be after original")
		}
	})

	t.Run("non-existent wallet", func(t *testing.T) {
		_, err := repo.UpdateDescription(ctx, "00000000-0000-0000-0000-000000000000", "x")
		if err == nil {
			t.Fatal("expected error for non-existent wallet, got nil")
		}
	})
}
