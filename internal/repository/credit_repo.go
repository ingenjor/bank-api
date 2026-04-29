package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"bank-api/internal/models"
)

type CreditRepo struct{ db *sql.DB }

func NewCreditRepo(db *sql.DB) *CreditRepo { return &CreditRepo{db} }

func (r *CreditRepo) Create(ctx context.Context, c *models.Credit) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO credits (id, user_id, amount, rate, term_months, monthly_payment, remaining, next_payment_date, status) 
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		c.ID, c.UserID, c.Amount, c.Rate, c.TermMonths, c.MonthlyPayment, c.Remaining, c.NextPaymentDate, c.Status)
	return err
}

func (r *CreditRepo) GetByID(ctx context.Context, id string) (*models.Credit, error) {
	c := &models.Credit{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, amount, rate, term_months, monthly_payment, remaining, next_payment_date, status FROM credits WHERE id=$1`, id).
		Scan(&c.ID, &c.UserID, &c.Amount, &c.Rate, &c.TermMonths, &c.MonthlyPayment, &c.Remaining, &c.NextPaymentDate, &c.Status)
	return c, err
}

func (r *CreditRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Credit, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, amount, rate, term_months, monthly_payment, remaining, next_payment_date, status FROM credits WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var credits []*models.Credit
	for rows.Next() {
		c := &models.Credit{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.Amount, &c.Rate, &c.TermMonths, &c.MonthlyPayment, &c.Remaining, &c.NextPaymentDate, &c.Status); err != nil {
			return nil, err
		}
		credits = append(credits, c)
	}
	return credits, nil
}

func (r *CreditRepo) AddSchedule(ctx context.Context, creditID string, schedule []models.PaymentScheduleItem) error {
	for _, item := range schedule {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO payment_schedules (credit_id, due_date, amount, status) VALUES ($1,$2,$3,'pending')`,
			creditID, item.DueDate, item.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CreditRepo) GetSchedule(ctx context.Context, creditID string) ([]models.PaymentScheduleItem, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT due_date, amount, status, COALESCE(penalty_applied, false) FROM payment_schedules WHERE credit_id=$1 ORDER BY due_date`, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.PaymentScheduleItem
	for rows.Next() {
		item := models.PaymentScheduleItem{}
		if err := rows.Scan(&item.DueDate, &item.Amount, &item.Status, &item.PenaltyApplied); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *CreditRepo) GetOverduePayments(ctx context.Context) ([]*models.Credit, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT c.id, c.user_id, c.amount, c.rate, c.term_months, c.monthly_payment, c.remaining, c.next_payment_date, c.status 
         FROM credits c 
         WHERE c.next_payment_date <= $1 AND c.status='active'`, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var credits []*models.Credit
	for rows.Next() {
		c := &models.Credit{}
		if err := rows.Scan(&c.ID, &c.UserID, &c.Amount, &c.Rate, &c.TermMonths, &c.MonthlyPayment, &c.Remaining, &c.NextPaymentDate, &c.Status); err != nil {
			return nil, err
		}
		credits = append(credits, c)
	}
	return credits, nil
}

func (r *CreditRepo) Update(ctx context.Context, c *models.Credit) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE credits SET remaining=$1, next_payment_date=$2, status=$3 WHERE id=$4`,
		c.Remaining, c.NextPaymentDate, c.Status, c.ID)
	return err
}

func (r *CreditRepo) MarkPaymentAsPaid(ctx context.Context, creditID string, dueDate time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE payment_schedules SET status='paid', paid_at=NOW() WHERE credit_id=$1 AND due_date=$2`,
		creditID, dueDate)
	return err
}

func (r *CreditRepo) HasPenaltyBeenApplied(ctx context.Context, creditID string, dueDate time.Time) (bool, error) {
	var applied bool
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(penalty_applied, false) FROM payment_schedules WHERE credit_id=$1 AND due_date=$2`,
		creditID, dueDate).Scan(&applied)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return applied, nil
}

func (r *CreditRepo) ApplyPenalty(ctx context.Context, creditID string, dueDate time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE payment_schedules SET penalty_applied = true WHERE credit_id=$1 AND due_date=$2`,
		creditID, dueDate)
	return err
}
