package scheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bank-api/internal/integration"
	"bank-api/internal/repository"
	"bank-api/internal/scheduler"
	"bank-api/internal/service"
)

type emailSenderMock struct{ called bool }

func (e *emailSenderMock) Send(to, subject, body string) error {
	e.called = true
	return nil
}

func TestPaymentScheduler_ProcessOverdueAndSendsEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	creditRepo := repository.NewCreditRepo(db)
	accountRepo := repository.NewAccountRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)
	userRepo := repository.NewUserRepo(db)

	// Подменяем URL ЦБ, чтобы не делать реальные запросы.
	integration.CBRServiceURL = "http://127.0.0.1:1"
	defer func() {
		integration.CBRServiceURL = "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx"
	}()

	cbrClient := integration.NewCBRClient()
	creditService := service.NewCreditService(creditRepo, accountRepo, transactionRepo, cbrClient)

	now := time.Now()
	overdueDate := now.AddDate(0, -1, 0)
	overdueRows := sqlmock.NewRows([]string{"id", "user_id", "amount", "rate", "term_months", "monthly_payment", "remaining", "next_payment_date", "status"}).
		AddRow("cr1", "user1", decimal.NewFromInt(10000), decimal.NewFromFloat(10.0), 6, decimal.NewFromInt(2000), decimal.NewFromInt(8000), overdueDate, "active")
	mock.ExpectQuery("SELECT .* FROM credits c WHERE c.next_payment_date <= \\$1 AND c.status='active'").
		WithArgs(sqlmock.AnyArg()).WillReturnRows(overdueRows)

	accRows := sqlmock.NewRows([]string{"id", "user_id", "balance", "created_at"}).
		AddRow("acc1", "user1", decimal.NewFromInt(5000), now)
	mock.ExpectQuery("SELECT id, user_id, balance, created_at FROM accounts WHERE user_id=\\$1").
		WithArgs("user1").WillReturnRows(accRows)

	mock.ExpectQuery("SELECT COALESCE\\(penalty_applied, false\\) FROM payment_schedules WHERE credit_id=\\$1 AND due_date=\\$2").
		WithArgs("cr1", sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"penalty_applied"}).AddRow(false))

	mock.ExpectExec("UPDATE accounts SET balance=\\$1 WHERE id=\\$2").
		WithArgs(decimal.NewFromInt(3000), "acc1").WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("INSERT INTO transactions").
		WithArgs("acc1", nil, decimal.NewFromInt(2000), "credit_payment", "cr1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE payment_schedules SET status='paid', paid_at=NOW\\(\\) WHERE credit_id=\\$1 AND due_date=\\$2").
		WithArgs("cr1", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("UPDATE credits SET remaining=\\$1, next_payment_date=\\$2, status=\\$3 WHERE id=\\$4").
		WithArgs(decimal.NewFromInt(6000), sqlmock.AnyArg(), "active", "cr1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	userRows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "created_at"}).
		AddRow("user1", "u", "user1@x.com", "hash", now)
	mock.ExpectQuery("SELECT id, username, email, password_hash, created_at FROM users WHERE id=\\$1").
		WithArgs("user1").WillReturnRows(userRows)

	logger := logrus.New()
	emailSender := &emailSenderMock{}

	sched := scheduler.NewPaymentScheduler(creditService, userRepo, emailSender, 50*time.Millisecond, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sched.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	cancel()

	assert.NoError(t, mock.ExpectationsWereMet())
	assert.True(t, emailSender.called, "expected email sender to be called")
}
