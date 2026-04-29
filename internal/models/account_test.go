package models_test

import (
	"testing"

	"bank-api/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestDepositRequestValidation(t *testing.T) {
	v := validator.New()
	req := models.DepositRequest{Amount: 100}
	assert.NoError(t, v.Struct(req))
	req.Amount = 0
	assert.Error(t, v.Struct(req))
}
