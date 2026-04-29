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

type mockTxRepoForTransfer struct {
	mock.Mock
}

func (m *mockTxRepoForTransfer) Record(ctx context.Context, fromAcc, toAcc string, amount decimal.Decimal, txType string) error {
	return nil
}
func (m *mockTxRepoForTransfer) RecordCreditTransaction(_ context.Context, _ string, _ string, _ decimal.Decimal, _ string, _ string) error {
	return nil
}
func (m *mockTxRepoForTransfer) GetMonthlySummary(_ context.Context, _ string, _ time.Time, _ time.Time) (decimal.Decimal, decimal.Decimal, error) {
	return decimal.Zero, decimal.Zero, nil
}
func (m *mockTxRepoForTransfer) GetUpcomingCreditPayments(_ context.Context, _ string, _ time.Time, _ time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}
func (m *mockTxRepoForTransfer) GetTotalPenaltiesForUser(_ context.Context, _ string, _ time.Time, _ time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

type mockAccRepoForTransfer struct {
	mock.Mock
}

func (m *mockAccRepoForTransfer) GetByID(ctx context.Context, id string) (*models.Account, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Account), args.Error(1)
}
func (m *mockAccRepoForTransfer) TransferTx(ctx context.Context, fromID, toID string, amount decimal.Decimal) error {
	args := m.Called(ctx, fromID, toID, amount)
	return args.Error(0)
}
func (m *mockAccRepoForTransfer) Create(_ context.Context, _ *models.Account) error { return nil }
func (m *mockAccRepoForTransfer) GetByUserID(_ context.Context, _ string) ([]*models.Account, error) {
	return nil, nil
}
func (m *mockAccRepoForTransfer) UpdateBalance(_ context.Context, _ string, _ decimal.Decimal) error {
	return nil
}

type mockUserRepoForTransfer struct {
	mock.Mock
}

func (m *mockUserRepoForTransfer) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockUserRepoForTransfer) GetByEmail(_ context.Context, _ string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepoForTransfer) Create(_ context.Context, _ *models.User) error { return nil }
func (m *mockUserRepoForTransfer) IsUnique(_ context.Context, _ string, _ string) (bool, error) {
	return true, nil
}

func TestTransactionService_Transfer_Success(t *testing.T) {
	accRepo := new(mockAccRepoForTransfer)
	txRepo := new(mockTxRepoForTransfer)
	userRepo := new(mockUserRepoForTransfer)

	svc := service.NewTransactionService(txRepo, accRepo, nil, userRepo)

	fromAcc := &models.Account{ID: "from", UserID: "user1", Balance: decimal.NewFromInt(5000)}
	accRepo.On("GetByID", mock.Anything, "from").Return(fromAcc, nil)
	accRepo.On("TransferTx", mock.Anything, "from", "to", decimal.NewFromInt(500)).Return(nil)
	userRepo.On("GetByID", mock.Anything, "user1").Return(&models.User{ID: "user1", Email: "a@b.com"}, nil)

	err := svc.Transfer(context.Background(), "from", "to", "user1", decimal.NewFromInt(500))
	assert.NoError(t, err)
}
