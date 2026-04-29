package models_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"bank-api/internal/models"
)

func TestRegisterRequestValidation(t *testing.T) {
	v := validator.New()
	tests := []struct {
		name string
		req  models.RegisterRequest
		ok   bool
	}{
		{"valid", models.RegisterRequest{Username: "user1", Email: "a@b.com", Password: "password123"}, true},
		{"short username", models.RegisterRequest{Username: "ab", Email: "a@b.com", Password: "password123"}, false},
		{"invalid email", models.RegisterRequest{Username: "user1", Email: "bad", Password: "password123"}, false},
		{"short password", models.RegisterRequest{Username: "user1", Email: "a@b.com", Password: "123"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.req)
			if tt.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestLoginRequestValidation(t *testing.T) {
	v := validator.New()
	req := models.LoginRequest{Email: "a@b.com", Password: "pass"}
	assert.NoError(t, v.Struct(req))
	req.Email = ""
	assert.Error(t, v.Struct(req))
}
