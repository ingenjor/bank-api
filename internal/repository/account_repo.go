package repository

import (
	"bank-api/internal/models"
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
)

type AccountRepo struct{ db *sql.DB }

func NewAccountRepo(db *sql.DB) *AccountRepo { return &AccountRepo{db} }

func (r *AccountRepo) Create(ctx context.Context, a *models.Account) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO accounts (id, user_id, balance) VALUES ($1,$2,$3)`, a.ID, a.UserID, a.Balance)
	return err
}

func (r *AccountRepo) GetByID(ctx context.Context, id string) (*models.Account, error) {
	a := &models.Account{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, balance, created_at FROM accounts WHERE id=$1`, id).
		Scan(&a.ID, &a.UserID, &a.Balance, &a.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return a, nil
}

func (r *AccountRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Account, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, balance, created_at FROM accounts WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*models.Account
	for rows.Next() {
		a := &models.Account{}
		if err := rows.Scan(&a.ID, &a.UserID, &a.Balance, &a.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, a)
	}
	return res, rows.Err()
}

func (r *AccountRepo) UpdateBalance(ctx context.Context, id string, newBalance decimal.Decimal) error {
	_, err := r.db.ExecContext(ctx, `UPDATE accounts SET balance=$1 WHERE id=$2`, newBalance, id)
	return err
}

func (r *AccountRepo) TransferTx(ctx context.Context, fromID, toID string, amount decimal.Decimal) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM accounts WHERE id=$1)`, fromID).Scan(&exists)
	if err != nil || !exists {
		return fmt.Errorf("source account not found")
	}

	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM accounts WHERE id=$1)`, toID).Scan(&exists)
	if err != nil || !exists {
		return fmt.Errorf("recipient account not found")
	}

	result, err := tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance - $1 WHERE id=$2 AND balance >= $1`, amount, fromID)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("insufficient funds")
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE accounts SET balance = balance + $1 WHERE id=$2`, amount, toID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO transactions (from_account_id, to_account_id, amount, type) VALUES ($1,$2,$3,'transfer')`,
		fromID, toID, amount)
	if err != nil {
		return err
	}
	return tx.Commit()
}
