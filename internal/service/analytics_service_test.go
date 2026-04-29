package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bank-api/internal/repository"
	"bank-api/internal/service"
)

func TestAnalyticsService_PredictBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	txRepo := repository.NewTransactionRepo(db)
	accRepo := repository.NewAccountRepo(db)
	creditRepo := repository.NewCreditRepo(db)

	svc := service.NewAnalyticsService(txRepo, accRepo, creditRepo)

	// Мокаем GetByID
	rowsAcc := sqlmock.NewRows([]string{"id", "user_id", "balance", "created_at"}).
		AddRow("acc1", "user1", decimal.NewFromInt(10000), time.Now())
	mock.ExpectQuery("SELECT id, user_id, balance, created_at FROM accounts WHERE id=\\$1").
		WithArgs("acc1").WillReturnRows(rowsAcc)

	// Мокаем GetUpcomingCreditPayments (используем актуальный SQL с payment_schedules.amount)
	rowsPay := sqlmock.NewRows([]string{"total"}).AddRow(decimal.NewFromInt(2000))
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(payment_schedules.amount\\),0\\) FROM payment_schedules").
		WithArgs("user1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rowsPay)

	pred, err := svc.PredictBalance(context.Background(), "acc1", "user1", 30)
	assert.NoError(t, err)
	// Баланс 10000 – предстоящие платежи 2000 = 8000
	assert.True(t, decimal.NewFromInt(8000).Equal(pred), "expected 8000, got %s", pred)
}
