package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"bank-api/internal/repository"
)

type AnalyticsService struct {
	txRepo      *repository.TransactionRepo
	accountRepo *repository.AccountRepo
	creditRepo  *repository.CreditRepo
}

func NewAnalyticsService(tx *repository.TransactionRepo, acc *repository.AccountRepo, cr *repository.CreditRepo) *AnalyticsService {
	return &AnalyticsService{tx, acc, cr}
}

type AnalyticsData struct {
	MonthlyIncome     decimal.Decimal `json:"monthly_income"`
	MonthlyExpense    decimal.Decimal `json:"monthly_expense"`
	CreditLoad        decimal.Decimal `json:"credit_load"`
	BalancePrediction decimal.Decimal `json:"balance_prediction_30d"`
}

func (s *AnalyticsService) GetAnalytics(ctx context.Context, userID string) (*AnalyticsData, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	accounts, err := s.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var totalIncome, totalExpense decimal.Decimal
	for _, acc := range accounts {
		inc, exp, err := s.txRepo.GetMonthlySummary(ctx, acc.ID, monthStart, monthEnd)
		if err != nil {
			return nil, err
		}
		totalIncome = totalIncome.Add(inc)
		totalExpense = totalExpense.Add(exp)
	}

	penalties, err := s.txRepo.GetTotalPenaltiesForUser(ctx, userID, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	totalExpense = totalExpense.Add(penalties)

	credits, err := s.creditRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var load decimal.Decimal
	for _, c := range credits {
		if c.Status == "active" {
			load = load.Add(c.MonthlyPayment)
		}
	}

	totalBalance := decimal.Zero
	for _, acc := range accounts {
		totalBalance = totalBalance.Add(acc.Balance)
	}
	upcoming, err := s.txRepo.GetUpcomingCreditPayments(ctx, userID, now, now.AddDate(0, 0, 30))
	if err != nil {
		return nil, err
	}
	prediction := totalBalance.Sub(upcoming)

	return &AnalyticsData{
		MonthlyIncome:     totalIncome,
		MonthlyExpense:    totalExpense,
		CreditLoad:        load,
		BalancePrediction: prediction,
	}, nil
}

func (s *AnalyticsService) PredictBalance(ctx context.Context, accountID, userID string, days int) (decimal.Decimal, error) {
	acc, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return decimal.Zero, err
	}
	if acc.UserID != userID {
		return decimal.Zero, fmt.Errorf("access denied")
	}

	now := time.Now()
	upcoming, err := s.txRepo.GetUpcomingCreditPayments(ctx, userID, now, now.AddDate(0, 0, days))
	if err != nil {
		return decimal.Zero, err
	}
	predicted := acc.Balance.Sub(upcoming)
	return predicted, nil
}
