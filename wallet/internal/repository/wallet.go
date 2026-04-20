package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"wallet/internal/models"
)

type WalletRepository struct {
	dbPool *pgxpool.Pool
}

func NewWalletRepository(dbPool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{dbPool: dbPool}
}

func (r *WalletRepository) ListByUserID(ctx context.Context, userID string) ([]models.Wallet, error) {
	rows, err := r.dbPool.Query(ctx,
		`SELECT id, user_id, description, balance, created_at, updated_at FROM wallets WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list wallets: %w", err)
	}
	defer rows.Close()

	var wallets []models.Wallet
	for rows.Next() {
		var w models.Wallet
		var balance string
		err = rows.Scan(&w.ID, &w.UserID, &w.Description, &balance, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan wallet: %w", err)
		}
		w.Balance, err = decimal.NewFromString(balance)
		if err != nil {
			return nil, fmt.Errorf("parse balance: %w", err)
		}
		wallets = append(wallets, w)
	}
	return wallets, rows.Err()
}

func (r *WalletRepository) GetByID(ctx context.Context, id string) (*models.Wallet, error) {
	var w models.Wallet
	var balance string
	err := r.dbPool.QueryRow(ctx,
		`SELECT id, user_id, description, balance, created_at, updated_at FROM wallets WHERE id = $1`,
		id,
	).Scan(&w.ID, &w.UserID, &w.Description, &balance, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}
	w.Balance, err = decimal.NewFromString(balance)
	if err != nil {
		return nil, fmt.Errorf("parse balance: %w", err)
	}
	return &w, nil
}

func (r *WalletRepository) Create(ctx context.Context, userID, description string) (*models.Wallet, error) {
	now := time.Now().UTC()
	w := &models.Wallet{
		ID:          uuid.New().String(),
		UserID:      userID,
		Description: description,
		Balance:     decimal.Zero,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := r.dbPool.Exec(ctx,
		`INSERT INTO wallets (id, user_id, description, balance, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		w.ID, w.UserID, w.Description, w.Balance.String(), w.CreatedAt, w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return w, nil
}

func (r *WalletRepository) UpdateDescription(ctx context.Context, id, description string) (*models.Wallet, error) {
	now := time.Now().UTC()
	var w models.Wallet
	var balance string
	err := r.dbPool.QueryRow(ctx,
		`UPDATE wallets SET description = $1, updated_at = $2
		 WHERE id = $3
		 RETURNING id, user_id, description, balance, created_at, updated_at`,
		description, now, id,
	).Scan(&w.ID, &w.UserID, &w.Description, &balance, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update wallet: %w", err)
	}
	w.Balance, err = decimal.NewFromString(balance)
	if err != nil {
		return nil, fmt.Errorf("parse balance: %w", err)
	}
	return &w, nil
}
