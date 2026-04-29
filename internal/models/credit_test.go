package models_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"bank-api/internal/models"
)

func TestCreditApplicationRequestValidation(t *testing.T) {
	v := validator.New()
	req := models.CreditApplicationRequest{Amount: 10000, TermMonths: 12}
	assert.NoError(t, v.Struct(req))
	req.Amount = 0
	assert.Error(t, v.Struct(req))
	req.Amount = 1
	req.TermMonths = 0
	assert.Error(t, v.Struct(req))
	req.TermMonths = 61
	assert.Error(t, v.Struct(req))
}
