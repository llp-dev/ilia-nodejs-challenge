package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"wallet/internal/models"
)

var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrWalletNotFound = errors.New("wallet not found")

type TransactionRepository struct {
	dbPool *pgxpool.Pool
}

func NewTransactionRepository(dbPool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{dbPool: dbPool}
}

func (r *TransactionRepository) Create(ctx context.Context, walletID string, value decimal.Decimal, description, operationID string) (*models.Transaction, error) {
	tx, err := r.dbPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()

	tag, err := tx.Exec(ctx,
		`UPDATE wallets SET balance = balance + $1, updated_at = $2
		 WHERE id = $3 AND (balance + $1) >= 0`,
		value.String(), now, walletID,
	)
	if err != nil {
		return nil, fmt.Errorf("update wallet balance: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrInsufficientBalance
	}

	t := &models.Transaction{
		ID:          uuid.New().String(),
		WalletID:    walletID,
		Value:       value,
		Description: description,
		OperationID: operationID,
		CreatedAt:   now,
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO transactions (id, wallet_id, value, description, operation_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		t.ID, t.WalletID, t.Value.String(), t.Description, t.OperationID, t.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return t, nil
}
