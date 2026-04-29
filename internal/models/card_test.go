package models_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"bank-api/internal/models"
)

func TestCreateCardRequestValidation(t *testing.T) {
	v := validator.New()
	req := models.CreateCardRequest{AccountID: "550e8400-e29b-41d4-a716-446655440000"}
	assert.NoError(t, v.Struct(req))
	req.AccountID = "invalid-uuid"
	assert.Error(t, v.Struct(req))
}

func TestPaymentRequestValidation(t *testing.T) {
	v := validator.New()
	req := models.PaymentRequest{CardID: "550e8400-e29b-41d4-a716-446655440000", Amount: 10.5}
	assert.NoError(t, v.Struct(req))
	req.Amount = 0
	assert.Error(t, v.Struct(req))
	req.CardID = ""
	assert.Error(t, v.Struct(req))
}
