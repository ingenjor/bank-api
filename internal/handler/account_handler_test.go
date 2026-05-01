package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"bank-api/internal/handler"
	"bank-api/internal/middleware"
	"bank-api/internal/models"
)

type mockAccountService struct {
	mock.Mock
}

func (m *mockAccountService) Create(ctx context.Context, userID string) (*models.Account, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.Account), args.Error(1)
}
func (m *mockAccountService) GetUserAccounts(ctx context.Context, userID string) ([]*models.Account, error) {
	return nil, nil
}
func (m *mockAccountService) Deposit(ctx context.Context, accountID, userID string, amount decimal.Decimal) error {
	args := m.Called(ctx, accountID, userID, amount)
	return args.Error(0)
}
func (m *mockAccountService) Withdraw(ctx context.Context, accountID, userID string, amount decimal.Decimal) error {
	return nil
}

func TestAccountHandler_Deposit(t *testing.T) {
	mockSvc := new(mockAccountService)
	logger := logrus.New()
	h := handler.NewAccountHandler(mockSvc, logger)

	mockSvc.On("Deposit", mock.Anything, "acc1", "user1", mock.AnythingOfType("decimal.Decimal")).Return(nil)

	r := mux.NewRouter()
	r.HandleFunc("/accounts/{id}/deposit", h.Deposit).Methods("POST")

	body, _ := json.Marshal(map[string]float64{"amount": 500})
	req := httptest.NewRequest("POST", "/accounts/acc1/deposit", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "acc1"})
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user1")
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "deposit successful")
}
