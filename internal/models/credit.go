package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Credit struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	Amount          decimal.Decimal `json:"amount"`
	Rate            decimal.Decimal `json:"rate"`
	TermMonths      int             `json:"term_months"`
	MonthlyPayment  decimal.Decimal `json:"monthly_payment"`
	Remaining       decimal.Decimal `json:"remaining"`
	NextPaymentDate time.Time       `json:"next_payment_date"`
	Status          string          `json:"status"`
}

type CreditApplicationRequest struct {
	Amount     float64 `json:"amount"      validate:"required,gt=0"`
	TermMonths int     `json:"term_months" validate:"required,gte=1,lte=60"`
}

type PaymentScheduleItem struct {
	DueDate        time.Time       `json:"due_date"`
	Amount         decimal.Decimal `json:"amount"`
	Status         string          `json:"status"`
	PenaltyApplied bool            `json:"penalty_applied"`
}
