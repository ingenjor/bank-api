package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID            string          `json:"id"`
	FromAccountID string          `json:"from_account_id,omitempty"`
	ToAccountID   string          `json:"to_account_id,omitempty"`
	Amount        decimal.Decimal `json:"amount"`
	Type          string          `json:"type"`
	Timestamp     time.Time       `json:"timestamp"`
}
