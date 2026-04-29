package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"bank-api/internal/models"
	"bank-api/internal/repository"
)

func TestCreditRepo_Create(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewCreditRepo(db)
	c := &models.Credit{
		ID:              "cr1",
		UserID:          "u1",
		Amount:          decimal.NewFromInt(100000),
		Rate:            decimal.NewFromFloat(10.5),
		TermMonths:      12,
		MonthlyPayment:  decimal.NewFromInt(9000),
		Remaining:       decimal.NewFromInt(100000),
		NextPaymentDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Status:          "active",
	}
	mock.ExpectExec("INSERT INTO credits").
		WithArgs("cr1", "u1", decimal.NewFromInt(100000), decimal.NewFromFloat(10.5), 12, decimal.NewFromInt(9000), decimal.NewFromInt(100000), time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "active").
		WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.Create(context.Background(), c)
	assert.NoError(t, err)
}

func TestCreditRepo_GetOverduePayments(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewCreditRepo(db)
	rows := sqlmock.NewRows([]string{"id", "user_id", "amount", "rate", "term_months", "monthly_payment", "remaining", "next_payment_date", "status"}).
		AddRow("cr1", "u1", decimal.NewFromInt(10000), decimal.NewFromFloat(5.0), 6, decimal.NewFromInt(2000), decimal.NewFromInt(8000), time.Now().AddDate(0, -1, 0), "active")
	mock.ExpectQuery("SELECT .* FROM credits c WHERE c.next_payment_date <= \\$1 AND c.status='active'").
		WithArgs(sqlmock.AnyArg()).WillReturnRows(rows)
	credits, err := repo.GetOverduePayments(context.Background())
	assert.NoError(t, err)
	assert.Len(t, credits, 1)
}
