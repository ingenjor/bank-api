package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Account struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
}

type DepositRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}
