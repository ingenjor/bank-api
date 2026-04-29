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

func TestAccountRepo_Create(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewAccountRepo(db)
	a := &models.Account{ID: "acc1", UserID: "user1", Balance: decimal.NewFromInt(100)}
	mock.ExpectExec("INSERT INTO accounts").
		WithArgs("acc1", "user1", decimal.NewFromInt(100)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.Create(context.Background(), a)
	assert.NoError(t, err)
}

func TestAccountRepo_GetByID(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewAccountRepo(db)
	rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "created_at"}).
		AddRow("acc1", "user1", decimal.NewFromInt(200), time.Now())
	mock.ExpectQuery("SELECT id, user_id, balance, created_at FROM accounts WHERE id=\\$1").
		WithArgs("acc1").WillReturnRows(rows)
	acc, err := repo.GetByID(context.Background(), "acc1")
	assert.NoError(t, err)
	assert.Equal(t, "user1", acc.UserID)
}

func TestAccountRepo_TransferTx_Success(t *testing.T) {
	db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	defer db.Close()
	repo := repository.NewAccountRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM accounts WHERE id=$1)`).WithArgs("from").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`SELECT EXISTS(SELECT 1 FROM accounts WHERE id=$1)`).WithArgs("to").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`UPDATE accounts SET balance = balance - $1 WHERE id=$2 AND balance >= $1`).
		WithArgs(decimal.NewFromInt(100), "from").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE accounts SET balance = balance + $1 WHERE id=$2`).
		WithArgs(decimal.NewFromInt(100), "to").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO transactions (from_account_id, to_account_id, amount, type) VALUES ($1,$2,$3,'transfer')`).
		WithArgs("from", "to", decimal.NewFromInt(100)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.TransferTx(context.Background(), "from", "to", decimal.NewFromInt(100))
	assert.NoError(t, err)
}

func TestAccountRepo_TransferTx_RecipientNotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewAccountRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM accounts WHERE id=\\$1\\)").WithArgs("from").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM accounts WHERE id=\\$1\\)").WithArgs("to").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectRollback()

	err := repo.TransferTx(context.Background(), "from", "to", decimal.NewFromInt(100))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recipient account not found")
}
