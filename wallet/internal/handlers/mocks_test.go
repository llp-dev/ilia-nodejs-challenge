package handlers_test

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"wallet/internal/models"
)

var errDB = errors.New("connection refused")

// mockWalletRepo is a configurable stub for the walletRepository interface.
type mockWalletRepo struct {
	listFn              func(ctx context.Context) ([]models.Wallet, error)
	getByIDFn           func(ctx context.Context, id string) (*models.Wallet, error)
	createFn            func(ctx context.Context, userID, description string) (*models.Wallet, error)
	updateDescriptionFn func(ctx context.Context, id, description string) (*models.Wallet, error)
}

func (m *mockWalletRepo) List(ctx context.Context) ([]models.Wallet, error) {
	return m.listFn(ctx)
}

func (m *mockWalletRepo) GetByID(ctx context.Context, id string) (*models.Wallet, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockWalletRepo) Create(ctx context.Context, userID, description string) (*models.Wallet, error) {
	return m.createFn(ctx, userID, description)
}

func (m *mockWalletRepo) UpdateDescription(ctx context.Context, id, description string) (*models.Wallet, error) {
	return m.updateDescriptionFn(ctx, id, description)
}

// mockTransactionRepo is a configurable stub for the transactionRepository interface.
type mockTransactionRepo struct {
	createFn func(ctx context.Context, walletID string, value decimal.Decimal, description, operationID string) (*models.Transaction, error)
}

func (m *mockTransactionRepo) Create(ctx context.Context, walletID string, value decimal.Decimal, description, operationID string) (*models.Transaction, error) {
	return m.createFn(ctx, walletID, value, description, operationID)
}

// mockUsersClient is a configurable stub for the usersClient interface.
type mockUsersClient struct {
	getUserFn func(ctx context.Context, userID string) (string, error)
}

func (m *mockUsersClient) GetUser(ctx context.Context, userID string) (string, error) {
	return m.getUserFn(ctx, userID)
}
