package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"bank-api/internal/models"
)

type AccountRepository interface {
	Create(ctx context.Context, a *models.Account) error
	GetByID(ctx context.Context, id string) (*models.Account, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Account, error)
	UpdateBalance(ctx context.Context, id string, newBalance decimal.Decimal) error
	TransferTx(ctx context.Context, fromID, toID string, amount decimal.Decimal) error
}

type AccountService struct {
	repo   AccountRepository
	txRepo TransactionRepository
}

func NewAccountService(repo AccountRepository, txRepo TransactionRepository) *AccountService {
	return &AccountService{repo: repo, txRepo: txRepo}
}

func (s *AccountService) Create(ctx context.Context, userID string) (*models.Account, error) {
	acc := &models.Account{
		ID:        uuid.New().String(),
		UserID:    userID,
		Balance:   decimal.Zero,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *AccountService) GetUserAccounts(ctx context.Context, userID string) ([]*models.Account, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *AccountService) Deposit(ctx context.Context, accountID, userID string, amount decimal.Decimal) error {
	acc, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	if acc.UserID != userID {
		return errors.New("access denied")
	}
	newBalance := acc.Balance.Add(amount)
	if err := s.repo.UpdateBalance(ctx, accountID, newBalance); err != nil {
		return err
	}
	return s.txRepo.Record(ctx, "", accountID, amount, "deposit")
}

func (s *AccountService) Withdraw(ctx context.Context, accountID, userID string, amount decimal.Decimal) error {
	acc, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	if acc.UserID != userID {
		return errors.New("access denied")
	}
	if acc.Balance.LessThan(amount) {
		return errors.New("insufficient funds")
	}
	newBalance := acc.Balance.Sub(amount)
	if err := s.repo.UpdateBalance(ctx, accountID, newBalance); err != nil {
		return err
	}
	return s.txRepo.Record(ctx, accountID, "", amount, "withdrawal")
}
