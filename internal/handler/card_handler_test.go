package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"bank-api/internal/handler"
	"bank-api/internal/middleware"
	"bank-api/internal/models"
)

type mockCardService struct {
	mock.Mock
}

func (m *mockCardService) Issue(ctx context.Context, userID, accountID string) (*models.CardResponse, error) {
	args := m.Called(ctx, userID, accountID)
	return args.Get(0).(*models.CardResponse), args.Error(1)
}
func (m *mockCardService) GetByID(ctx context.Context, cardID, userID string) (*models.CardResponse, error) {
	return nil, nil
}
func (m *mockCardService) GetUserCards(ctx context.Context, userID string) ([]*models.CardResponse, error) {
	return nil, nil
}
func (m *mockCardService) Payment(ctx context.Context, userID string, req models.PaymentRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func TestCardHandler_Payment(t *testing.T) {
	mockSvc := new(mockCardService)
	logger := logrus.New()
	h := handler.NewCardHandler(mockSvc, logger)

	validCardID := "550e8400-e29b-41d4-a716-446655440000"
	mockSvc.On("Payment", mock.Anything, "user1", models.PaymentRequest{
		CardID: validCardID,
		Amount: 150.0,
	}).Return(nil)

	r := mux.NewRouter()
	r.HandleFunc("/cards/payment", h.Payment).Methods("POST")
	body, _ := json.Marshal(map[string]interface{}{"card_id": validCardID, "amount": 150.0})
	req := httptest.NewRequest("POST", "/cards/payment", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user1"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "payment successful")
}
