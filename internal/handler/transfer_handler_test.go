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
)

type mockTransactionService struct {
	mock.Mock
}

func (m *mockTransactionService) Transfer(ctx context.Context, fromID, toID, userID string, amount decimal.Decimal) error {
	args := m.Called(ctx, fromID, toID, userID, amount)
	return args.Error(0)
}

func TestTransferHandler_Transfer(t *testing.T) {
	mockSvc := new(mockTransactionService)
	logger := logrus.New()
	h := handler.NewTransferHandler(mockSvc, logger)

	validFrom := "550e8400-e29b-41d4-a716-446655440001"
	validTo := "550e8400-e29b-41d4-a716-446655440002"

	mockSvc.On("Transfer", mock.Anything, validFrom, validTo, "user1", mock.AnythingOfType("decimal.Decimal")).Return(nil)

	r := mux.NewRouter()
	r.HandleFunc("/transfer", h.Transfer).Methods("POST")
	body, _ := json.Marshal(map[string]interface{}{
		"from":   validFrom,
		"to":     validTo,
		"amount": 1000.0,
	})
	req := httptest.NewRequest("POST", "/transfer", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), "userID", "user1"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "transfer completed")
}
