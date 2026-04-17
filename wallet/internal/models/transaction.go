package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID          string          `json:"id"`
	WalletID    string          `json:"wallet_id"`
	Value       decimal.Decimal `json:"value"`
	Description string          `json:"description"`
	OperationID string          `json:"operation_id"`
	CreatedAt   time.Time       `json:"created_at"`
}
