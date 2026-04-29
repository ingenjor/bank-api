package repository

import (
	"bank-api/internal/models"
	"context"
	"database/sql"
)

type CardRepo struct{ db *sql.DB }

func NewCardRepo(db *sql.DB) *CardRepo { return &CardRepo{db} }

func (r *CardRepo) Create(ctx context.Context, card *models.Card) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO cards (id, account_id, encrypted_number, hmac_number, encrypted_expiry, cvv_hash, status, last_four) 
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		card.ID, card.AccountID, card.EncryptedNumber, card.HMACNumber, card.EncryptedExpiry, card.CVVHash, card.Status, card.LastFour)
	return err
}

func (r *CardRepo) GetByID(ctx context.Context, id string) (*models.Card, error) {
	c := &models.Card{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, account_id, encrypted_number, hmac_number, encrypted_expiry, cvv_hash, status, last_four FROM cards WHERE id=$1`, id).
		Scan(&c.ID, &c.AccountID, &c.EncryptedNumber, &c.HMACNumber, &c.EncryptedExpiry, &c.CVVHash, &c.Status, &c.LastFour)
	return c, err
}

func (r *CardRepo) GetByAccountID(ctx context.Context, accountID string) ([]*models.Card, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, account_id, encrypted_number, hmac_number, encrypted_expiry, cvv_hash, status, last_four FROM cards WHERE account_id=$1`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []*models.Card
	for rows.Next() {
		c := &models.Card{}
		if err := rows.Scan(&c.ID, &c.AccountID, &c.EncryptedNumber, &c.HMACNumber, &c.EncryptedExpiry, &c.CVVHash, &c.Status, &c.LastFour); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

func (r *CardRepo) GetCardsByUserID(ctx context.Context, userID string) ([]*models.Card, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT c.id, c.account_id, c.encrypted_number, c.hmac_number, c.encrypted_expiry, c.cvv_hash, c.status, c.last_four 
		 FROM cards c JOIN accounts a ON c.account_id = a.id 
		 WHERE a.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []*models.Card
	for rows.Next() {
		c := &models.Card{}
		if err := rows.Scan(&c.ID, &c.AccountID, &c.EncryptedNumber, &c.HMACNumber, &c.EncryptedExpiry, &c.CVVHash, &c.Status, &c.LastFour); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}
