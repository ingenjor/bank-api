package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"bank-api/internal/handler"
	"bank-api/internal/models"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, req models.RegisterRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}
func (m *mockAuthService) Login(ctx context.Context, req models.LoginRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockSvc := new(mockAuthService)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	h := handler.NewAuthHandler(mockSvc, logger)

	mockSvc.On("Register", mock.Anything, models.RegisterRequest{
		Username: "validuser",
		Email:    "u@x.com",
		Password: "password123",
	}).Return(nil)

	r := mux.NewRouter()
	r.HandleFunc("/register", h.Register).Methods("POST")
	body, _ := json.Marshal(map[string]string{
		"username": "validuser",
		"email":    "u@x.com",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "registered successfully")
}

func TestAuthHandler_Login(t *testing.T) {
	mockSvc := new(mockAuthService)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	h := handler.NewAuthHandler(mockSvc, logger)

	mockSvc.On("Login", mock.Anything, models.LoginRequest{Email: "a@x.com", Password: "pass"}).Return("token123", nil)

	r := mux.NewRouter()
	r.HandleFunc("/login", h.Login).Methods("POST")
	body, _ := json.Marshal(map[string]string{"email": "a@x.com", "password": "pass"})
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "token123", resp["token"])
}
