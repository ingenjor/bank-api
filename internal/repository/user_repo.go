package repository

import (
	"bank-api/internal/models"
	"context"
	"database/sql"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db} }

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, username, email, password_hash) VALUES ($1,$2,$3,$4)`,
		u.ID, u.Username, u.Email, u.PasswordHash)
	return err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, created_at FROM users WHERE email=$1`, email).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, created_at FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (r *UserRepo) IsUnique(ctx context.Context, email, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email=$1 OR username=$2)`, email, username).Scan(&exists)
	return !exists, err
}
