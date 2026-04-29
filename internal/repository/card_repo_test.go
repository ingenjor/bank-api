package repository_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"bank-api/internal/models"
	"bank-api/internal/repository"
)

func TestCardRepo_Create(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewCardRepo(db)
	card := &models.Card{
		ID:              "c1",
		AccountID:       "acc1",
		EncryptedNumber: []byte("enc"),
		HMACNumber:      "hmac",
		EncryptedExpiry: []byte("encExp"),
		CVVHash:         "hash",
		Status:          "active",
		LastFour:        "1234",
	}
	mock.ExpectExec("INSERT INTO cards").
		WithArgs("c1", "acc1", []byte("enc"), "hmac", []byte("encExp"), "hash", "active", "1234").
		WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.Create(context.Background(), card)
	assert.NoError(t, err)
}

func TestCardRepo_GetByID(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewCardRepo(db)
	rows := sqlmock.NewRows([]string{"id", "account_id", "encrypted_number", "hmac_number", "encrypted_expiry", "cvv_hash", "status", "last_four"}).
		AddRow("c1", "acc1", []byte("enc"), "hmac", []byte("exp"), "hash", "active", "1234")
	mock.ExpectQuery("SELECT id, account_id, encrypted_number, hmac_number, encrypted_expiry, cvv_hash, status, last_four FROM cards WHERE id=\\$1").
		WithArgs("c1").WillReturnRows(rows)
	card, err := repo.GetByID(context.Background(), "c1")
	assert.NoError(t, err)
	assert.Equal(t, "acc1", card.AccountID)
}
