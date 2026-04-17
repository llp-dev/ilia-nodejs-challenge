package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Wallet struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id"`
	Description string          `json:"description"`
	Balance     decimal.Decimal `json:"balance"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
