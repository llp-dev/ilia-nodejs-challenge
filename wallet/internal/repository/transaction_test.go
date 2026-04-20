package repository_test

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"wallet/internal/repository"
	"wallet/internal/testhelper"
)

func setupWallet(t *testing.T, walletRepo *repository.WalletRepository, txRepo *repository.TransactionRepository, initialBalance string) string {
	t.Helper()
	ctx := context.Background()

	wallet, err := walletRepo.Create(ctx, "550e8400-e29b-41d4-a716-446655440000", "test wallet")
	if err != nil {
		t.Fatalf("create wallet: %v", err)
	}

	if initialBalance != "0" {
		_, err = txRepo.Create(ctx, wallet.ID, decimal.RequireFromString(initialBalance), "initial credit", "00000000-0000-0000-0000-000000000001")
		if err != nil {
			t.Fatalf("seed wallet balance: %v", err)
		}
	}

	return wallet.ID
}

func TestTransactionRepository_Create(t *testing.T) {
	testhelper.Truncate(t, testPool)
	walletRepo := repository.NewWalletRepository(testPool)
	txRepo := repository.NewTransactionRepository(testPool)
	ctx := context.Background()

	t.Run("credit increases balance", func(t *testing.T) {
		testhelper.Truncate(t, testPool)
		walletID := setupWallet(t, walletRepo, txRepo, "0")

		tx, err := txRepo.Create(ctx, walletID, decimal.NewFromInt(100), "credit", "00000000-0000-0000-0000-000000000002")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if tx.ID == "" {
			t.Error("expected non-empty transaction ID")
		}
		if !tx.Value.Equal(decimal.NewFromInt(100)) {
			t.Errorf("Value = %v, want 100", tx.Value)
		}

		wallet, err := walletRepo.GetByID(ctx, walletID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if !wallet.Balance.Equal(decimal.NewFromInt(100)) {
			t.Errorf("Balance = %v, want 100", wallet.Balance)
		}
	})

	t.Run("debit decreases balance", func(t *testing.T) {
		testhelper.Truncate(t, testPool)
		walletID := setupWallet(t, walletRepo, txRepo, "100")

		_, err := txRepo.Create(ctx, walletID, decimal.NewFromInt(-40), "debit", "00000000-0000-0000-0000-000000000003")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		wallet, err := walletRepo.GetByID(ctx, walletID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if !wallet.Balance.Equal(decimal.NewFromInt(60)) {
			t.Errorf("Balance = %v, want 60", wallet.Balance)
		}
	})

	t.Run("debit fails on insufficient balance", func(t *testing.T) {
		testhelper.Truncate(t, testPool)
		walletID := setupWallet(t, walletRepo, txRepo, "50")

		_, err := txRepo.Create(ctx, walletID, decimal.NewFromInt(-100), "overdraft", "00000000-0000-0000-0000-000000000004")
		if err == nil {
			t.Fatal("expected ErrInsufficientBalance, got nil")
		}
		if err != repository.ErrInsufficientBalance {
			t.Errorf("error = %v, want ErrInsufficientBalance", err)
		}

		wallet, err := walletRepo.GetByID(ctx, walletID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if !wallet.Balance.Equal(decimal.NewFromInt(50)) {
			t.Errorf("Balance = %v after failed debit, want 50", wallet.Balance)
		}
	})

	t.Run("fails on non-existent wallet", func(t *testing.T) {
		_, err := txRepo.Create(ctx, "00000000-0000-0000-0000-000000000000", decimal.NewFromInt(10), "ghost", "00000000-0000-0000-0000-000000000005")
		if err == nil {
			t.Fatal("expected error for non-existent wallet, got nil")
		}
	})

	t.Run("fails on duplicate operation_id", func(t *testing.T) {
		testhelper.Truncate(t, testPool)
		walletID := setupWallet(t, walletRepo, txRepo, "100")

		_, err := txRepo.Create(ctx, walletID, decimal.NewFromInt(10), "first", "00000000-0000-0000-0000-000000000006")
		if err != nil {
			t.Fatalf("first Create() error = %v", err)
		}

		_, err = txRepo.Create(ctx, walletID, decimal.NewFromInt(10), "duplicate", "00000000-0000-0000-0000-000000000006")
		if err != repository.ErrDuplicateOperation {
			t.Errorf("error = %v, want ErrDuplicateOperation", err)
		}
	})
}
