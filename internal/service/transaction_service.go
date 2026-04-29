package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"bank-api/internal/integration"
	"bank-api/internal/models"
)

type TransactionRepository interface {
	Record(ctx context.Context, fromAcc, toAcc string, amount decimal.Decimal, txType string) error
	RecordCreditTransaction(ctx context.Context, fromAcc, toAcc string, amount decimal.Decimal, txType string, creditID string) error
	GetMonthlySummary(ctx context.Context, accountID string, monthStart, monthEnd time.Time) (income, expense decimal.Decimal, err error)
	GetUpcomingCreditPayments(ctx context.Context, userID string, from, to time.Time) (decimal.Decimal, error)
	GetTotalPenaltiesForUser(ctx context.Context, userID string, start, end time.Time) (decimal.Decimal, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
}

type TransactionService struct {
	txRepo      TransactionRepository
	accountRepo AccountRepository
	emailSender integration.EmailSender
	userRepo    UserRepository
}

func NewTransactionService(tx TransactionRepository, acc AccountRepository, es integration.EmailSender, ur UserRepository) *TransactionService {
	return &TransactionService{txRepo: tx, accountRepo: acc, emailSender: es, userRepo: ur}
}

func (s *TransactionService) Transfer(ctx context.Context, fromID, toID, userID string, amount decimal.Decimal) error {
	fromAcc, err := s.accountRepo.GetByID(ctx, fromID)
	if err != nil || fromAcc.UserID != userID {
		return errors.New("access denied or account not found")
	}
	if fromAcc.Balance.LessThan(amount) {
		return errors.New("insufficient funds")
	}
	if err := s.accountRepo.TransferTx(ctx, fromID, toID, amount); err != nil {
		return err
	}
	if s.emailSender != nil && s.userRepo != nil {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err == nil && user != nil {
			body := fmt.Sprintf("Перевод %.2f RUB со счета %s на счет %s выполнен.", amount.InexactFloat64(), fromID, toID)
			go s.emailSender.Send(user.Email, "Перевод выполнен", body)
		}
	}
	return nil
}
