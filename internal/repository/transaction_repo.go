package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

type TransactionRepo struct{ db *sql.DB }

func NewTransactionRepo(db *sql.DB) *TransactionRepo { return &TransactionRepo{db} }

func (r *TransactionRepo) Record(ctx context.Context, fromAcc, toAcc string, amount decimal.Decimal, txType string) error {
	var fromPtr, toPtr interface{}
	if fromAcc != "" {
		fromPtr = fromAcc
	}
	if toAcc != "" {
		toPtr = toAcc
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO transactions (from_account_id, to_account_id, amount, type) VALUES ($1,$2,$3,$4)`,
		fromPtr, toPtr, amount, txType)
	return err
}

func (r *TransactionRepo) RecordCreditTransaction(ctx context.Context, fromAcc, toAcc string, amount decimal.Decimal, txType string, creditID string) error {
	var fromPtr, toPtr interface{}
	if fromAcc != "" {
		fromPtr = fromAcc
	}
	if toAcc != "" {
		toPtr = toAcc
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO transactions (from_account_id, to_account_id, amount, type, credit_id) VALUES ($1,$2,$3,$4,$5)`,
		fromPtr, toPtr, amount, txType, creditID)
	return err
}

func (r *TransactionRepo) GetMonthlySummary(ctx context.Context, accountID string, monthStart, monthEnd time.Time) (income, expense decimal.Decimal, err error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount),0) FROM transactions WHERE to_account_id=$1 AND created_at BETWEEN $2 AND $3 AND type IN ('deposit','transfer')`,
		accountID, monthStart, monthEnd)
	if err := row.Scan(&income); err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	row = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount),0) FROM transactions WHERE from_account_id=$1 AND created_at BETWEEN $2 AND $3 AND type IN ('withdrawal','transfer','payment','credit_payment')`,
		accountID, monthStart, monthEnd)
	if err := row.Scan(&expense); err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	return income, expense, nil
}

func (r *TransactionRepo) GetUpcomingCreditPayments(ctx context.Context, userID string, from, to time.Time) (decimal.Decimal, error) {
	var total decimal.Decimal
	row := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(payment_schedules.amount),0) FROM payment_schedules 
		 JOIN credits ON payment_schedules.credit_id = credits.id
		 WHERE credits.user_id=$1 AND payment_schedules.due_date BETWEEN $2 AND $3 AND payment_schedules.status='pending'`,
		userID, from, to)
	if err := row.Scan(&total); err != nil {
		return decimal.Zero, err
	}
	return total, nil
}

func (r *TransactionRepo) GetTotalPenaltiesForUser(ctx context.Context, userID string, start, end time.Time) (decimal.Decimal, error) {
	var total decimal.Decimal
	row := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(t.amount),0) FROM transactions t 
		 JOIN credits c ON t.credit_id = c.id 
		 WHERE c.user_id=$1 AND t.type='penalty' AND t.created_at BETWEEN $2 AND $3`,
		userID, start, end)
	if err := row.Scan(&total); err != nil {
		return decimal.Zero, err
	}
	return total, nil
}
