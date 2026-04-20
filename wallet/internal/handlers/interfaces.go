package handlers

import (
	"context"

	"github.com/shopspring/decimal"
	"wallet/internal/models"
)

type walletRepository interface {
	ListByUserID(ctx context.Context, userID string) ([]models.Wallet, error)
	GetByID(ctx context.Context, id string) (*models.Wallet, error)
	Create(ctx context.Context, userID, description string) (*models.Wallet, error)
	UpdateDescription(ctx context.Context, id, description string) (*models.Wallet, error)
}

type transactionRepository interface {
	Create(ctx context.Context, walletID string, value decimal.Decimal, description, operationID string) (*models.Transaction, error)
}

type usersClient interface {
	GetUser(ctx context.Context, userID string) (string, error)
}
