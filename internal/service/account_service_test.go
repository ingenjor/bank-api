package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"bank-api/internal/models"
	"bank-api/internal/service"
)

type mockAccountRepo struct {
	mock.Mock
}
type mockTxRepo struct {
	mock.Mock
}

func (m *mockAccountRepo) Create(ctx context.Context, a *models.Account) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}
func (m *mockAccountRepo) GetByID(ctx context.Context, id string) (*models.Account, error) {
	args := m.Called(ctx, id)
	acc := args.Get(0).(*models.Account)
	return acc, args.Error(1)
}
func (m *mockAccountRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Account, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.Account), args.Error(1)
}
func (m *mockAccountRepo) UpdateBalance(ctx context.Context, id string, newBalance decimal.Decimal) error {
	args := m.Called(ctx, id, newBalance)
	return args.Error(0)
}
func (m *mockAccountRepo) TransferTx(ctx context.Context, fromID, toID string, amount decimal.Decimal) error {
	args := m.Called(ctx, fromID, toID, amount)
	return args.Error(0)
}

func (m *mockTxRepo) Record(ctx context.Context, fromAcc, toAcc string, amount decimal.Decimal, txType string) error {
	args := m.Called(ctx, fromAcc, toAcc, amount, txType)
	return args.Error(0)
}
func (m *mockTxRepo) RecordCreditTransaction(ctx context.Context, fromAcc, toAcc string, amount decimal.Decimal, txType string, creditID string) error {
	return nil
}
func (m *mockTxRepo) GetMonthlySummary(ctx context.Context, accountID string, monthStart, monthEnd time.Time) (income, expense decimal.Decimal, err error) {
	return decimal.Zero, decimal.Zero, nil
}
func (m *mockTxRepo) GetUpcomingCreditPayments(ctx context.Context, userID string, from, to time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}
func (m *mockTxRepo) GetTotalPenaltiesForUser(ctx context.Context, userID string, start, end time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func TestAccountService_Deposit(t *testing.T) {
	accRepo := new(mockAccountRepo)
	txRepo := new(mockTxRepo)
	svc := service.NewAccountService(accRepo, txRepo)

	acc := &models.Account{ID: "acc1", UserID: "user1", Balance: decimal.NewFromInt(100)}
	accRepo.On("GetByID", mock.Anything, "acc1").Return(acc, nil)
	accRepo.On("UpdateBalance", mock.Anything, "acc1", decimal.NewFromInt(600)).Return(nil)
	txRepo.On("Record", mock.Anything, "", "acc1", decimal.NewFromInt(500), "deposit").Return(nil)

	err := svc.Deposit(context.Background(), "acc1", "user1", decimal.NewFromInt(500))
	assert.NoError(t, err)
}

func TestAccountService_Deposit_AccessDenied(t *testing.T) {
	accRepo := new(mockAccountRepo)
	svc := service.NewAccountService(accRepo, nil)
	acc := &models.Account{ID: "acc1", UserID: "otherUser", Balance: decimal.Zero}
	accRepo.On("GetByID", mock.Anything, "acc1").Return(acc, nil)
	err := svc.Deposit(context.Background(), "acc1", "user1", decimal.NewFromInt(100))
	assert.Error(t, err)
	assert.Equal(t, "access denied", err.Error())
}
