package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"bank-api/internal/repository"
)

func TestTransactionRepo_Record(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewTransactionRepo(db)
	mock.ExpectExec("INSERT INTO transactions \\(from_account_id, to_account_id, amount, type\\) VALUES \\(\\$1,\\$2,\\$3,\\$4\\)").
		WithArgs("from", nil, decimal.NewFromInt(500), "withdrawal").
		WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.Record(context.Background(), "from", "", decimal.NewFromInt(500), "withdrawal")
	assert.NoError(t, err)
}

func TestTransactionRepo_GetMonthlySummary(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewTransactionRepo(db)
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

	incomeRow := sqlmock.NewRows([]string{"sum"}).AddRow(decimal.NewFromInt(1500))
	expenseRow := sqlmock.NewRows([]string{"sum"}).AddRow(decimal.NewFromInt(500))

	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\),0\\) FROM transactions WHERE to_account_id=\\$1 AND created_at BETWEEN \\$2 AND \\$3 AND type IN \\('deposit','transfer'\\)").
		WithArgs("acc1", start, end).WillReturnRows(incomeRow)
	mock.ExpectQuery("SELECT COALESCE\\(SUM\\(amount\\),0\\) FROM transactions WHERE from_account_id=\\$1 AND created_at BETWEEN \\$2 AND \\$3 AND type IN \\('withdrawal','transfer','payment','credit_payment'\\)").
		WithArgs("acc1", start, end).WillReturnRows(expenseRow)

	inc, exp, err := repo.GetMonthlySummary(context.Background(), "acc1", start, end)
	assert.NoError(t, err)
	assert.Equal(t, decimal.NewFromInt(1500), inc)
	assert.Equal(t, decimal.NewFromInt(500), exp)
}
