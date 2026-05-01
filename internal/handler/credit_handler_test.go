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

type mockCreditService struct {
	mock.Mock
}

func (m *mockCreditService) Apply(ctx context.Context, userID string, req models.CreditApplicationRequest) (*models.Credit, error) {
	args := m.Called(ctx, userID, req)
	return args.Get(0).(*models.Credit), args.Error(1)
}
func (m *mockCreditService) GetSchedule(ctx context.Context, creditID, userID string) ([]models.PaymentScheduleItem, error) {
	return nil, nil
}

func TestCreditHandler_Apply(t *testing.T) {
	mockSvc := new(mockCreditService)
	logger := logrus.New()
	h := handler.NewCreditHandler(mockSvc, logger)

	credit := &models.Credit{ID: "cr1", Status: "active"}
	mockSvc.On("Apply", mock.Anything, "user1", models.CreditApplicationRequest{Amount: 100000, TermMonths: 12}).Return(credit, nil)

	r := mux.NewRouter()
	r.HandleFunc("/credits", h.Apply).Methods("POST")
	body, _ := json.Marshal(map[string]interface{}{"amount": 100000, "term_months": 12})
	req := httptest.NewRequest("POST", "/credits", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user1"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "active")
}
