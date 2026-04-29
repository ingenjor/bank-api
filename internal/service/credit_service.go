package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"bank-api/internal/integration"
	"bank-api/internal/models"
)

type CreditRepository interface {
	Create(ctx context.Context, c *models.Credit) error
	GetByID(ctx context.Context, id string) (*models.Credit, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Credit, error)
	AddSchedule(ctx context.Context, creditID string, schedule []models.PaymentScheduleItem) error
	GetSchedule(ctx context.Context, creditID string) ([]models.PaymentScheduleItem, error)
	GetOverduePayments(ctx context.Context) ([]*models.Credit, error)
	Update(ctx context.Context, c *models.Credit) error
	MarkPaymentAsPaid(ctx context.Context, creditID string, dueDate time.Time) error
	HasPenaltyBeenApplied(ctx context.Context, creditID string, dueDate time.Time) (bool, error)
	ApplyPenalty(ctx context.Context, creditID string, dueDate time.Time) error
}

type CreditService struct {
	creditRepo      CreditRepository
	accountRepo     AccountRepository
	transactionRepo TransactionRepository
	cbrClient       *integration.CBRClient
}

func NewCreditService(cr CreditRepository, ar AccountRepository, tr TransactionRepository, cbr *integration.CBRClient) *CreditService {
	return &CreditService{creditRepo: cr, accountRepo: ar, transactionRepo: tr, cbrClient: cbr}
}

func (s *CreditService) Apply(ctx context.Context, userID string, req models.CreditApplicationRequest) (*models.Credit, error) {
	keyRate, err := s.cbrClient.GetKeyRate()
	if err != nil {
		return nil, fmt.Errorf("cannot get key rate: %w", err)
	}
	annualRate := decimal.NewFromFloat(keyRate)
	monthlyRate := annualRate.Div(decimal.NewFromInt(12 * 100))
	amount := decimal.NewFromFloat(req.Amount)
	term := decimal.NewFromInt(int64(req.TermMonths))

	onePlusR := monthlyRate.Add(decimal.NewFromInt(1))
	pow := onePlusR.Pow(term)
	monthlyPayment := amount.Mul(monthlyRate).Mul(pow).Div(pow.Sub(decimal.NewFromInt(1)))

	credit := &models.Credit{
		ID:              uuid.New().String(),
		UserID:          userID,
		Amount:          amount,
		Rate:            annualRate,
		TermMonths:      req.TermMonths,
		MonthlyPayment:  monthlyPayment,
		Remaining:       amount,
		NextPaymentDate: time.Now().AddDate(0, 1, 0),
		Status:          "active",
	}
	if err := s.creditRepo.Create(ctx, credit); err != nil {
		return nil, err
	}
	schedule := make([]models.PaymentScheduleItem, req.TermMonths)
	for i := 0; i < req.TermMonths; i++ {
		due := time.Now().AddDate(0, i+1, 0)
		schedule[i] = models.PaymentScheduleItem{DueDate: due, Amount: monthlyPayment}
	}
	if err := s.creditRepo.AddSchedule(ctx, credit.ID, schedule); err != nil {
		return nil, err
	}
	return credit, nil
}

func (s *CreditService) GetSchedule(ctx context.Context, creditID, userID string) ([]models.PaymentScheduleItem, error) {
	credit, err := s.creditRepo.GetByID(ctx, creditID)
	if err != nil || credit.UserID != userID {
		return nil, errors.New("access denied")
	}
	return s.creditRepo.GetSchedule(ctx, creditID)
}

func (s *CreditService) ProcessOverdue(ctx context.Context) (map[string][]string, error) {
	overdue, err := s.creditRepo.GetOverduePayments(ctx)
	if err != nil {
		return nil, err
	}
	notifications := make(map[string][]string)

	for _, credit := range overdue {
		accounts, err := s.accountRepo.GetByUserID(ctx, credit.UserID)
		if err != nil || len(accounts) == 0 {
			continue
		}

		dueDate := credit.NextPaymentDate
		penaltyAlready, err := s.creditRepo.HasPenaltyBeenApplied(ctx, credit.ID, dueDate)
		if err != nil {
			continue
		}

		paid := false
		for _, acc := range accounts {
			if acc.Balance.GreaterThanOrEqual(credit.MonthlyPayment) {
				newBal := acc.Balance.Sub(credit.MonthlyPayment)
				if err := s.accountRepo.UpdateBalance(ctx, acc.ID, newBal); err == nil {
					err := s.transactionRepo.RecordCreditTransaction(ctx, acc.ID, "", credit.MonthlyPayment, "credit_payment", credit.ID)
					if err != nil {
						continue
					}
					s.creditRepo.MarkPaymentAsPaid(ctx, credit.ID, dueDate)
					credit.Remaining = credit.Remaining.Sub(credit.MonthlyPayment)
					credit.NextPaymentDate = credit.NextPaymentDate.AddDate(0, 1, 0)
					if credit.Remaining.LessThanOrEqual(decimal.Zero) {
						credit.Status = "paid"
					}
					s.creditRepo.Update(ctx, credit)
					notifications[credit.UserID] = append(notifications[credit.UserID],
						fmt.Sprintf("Credit %s: payment %.2f processed. Remaining: %.2f",
							credit.ID, credit.MonthlyPayment.InexactFloat64(), credit.Remaining.InexactFloat64()))
					paid = true
					break
				}
			}
		}
		if !paid {
			if !penaltyAlready {
				penalty := credit.MonthlyPayment.Mul(decimal.NewFromFloat(0.1))
				credit.Remaining = credit.Remaining.Add(penalty)
				s.transactionRepo.RecordCreditTransaction(ctx, "", "", penalty, "penalty", credit.ID)
				s.creditRepo.ApplyPenalty(ctx, credit.ID, dueDate)
				s.creditRepo.Update(ctx, credit)
				notifications[credit.UserID] = append(notifications[credit.UserID],
					fmt.Sprintf("Credit %s: overdue. Penalty %.2f applied. New remaining: %.2f",
						credit.ID, penalty.InexactFloat64(), credit.Remaining.InexactFloat64()))
			} else {
				notifications[credit.UserID] = append(notifications[credit.UserID],
					fmt.Sprintf("Credit %s: payment still overdue.", credit.ID))
			}
		}
	}
	return notifications, nil
}
