package handler_test

import (
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
	"bank-api/internal/service"
)

type mockAnalyticsService struct {
	mock.Mock
}

func (m *mockAnalyticsService) GetAnalytics(ctx context.Context, userID string) (*service.AnalyticsData, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*service.AnalyticsData), args.Error(1)
}
func (m *mockAnalyticsService) PredictBalance(ctx context.Context, accountID, userID string, days int) (decimal.Decimal, error) {
	args := m.Called(ctx, accountID, userID, days)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func TestAnalyticsHandler_PredictBalance(t *testing.T) {
	mockSvc := new(mockAnalyticsService)
	logger := logrus.New()
	h := handler.NewAnalyticsHandler(mockSvc, logger)

	mockSvc.On("PredictBalance", mock.Anything, "acc1", "user1", 90).Return(decimal.NewFromInt(5000), nil)

	r := mux.NewRouter()
	r.HandleFunc("/accounts/{accountId}/predict", h.PredictBalance).Methods("GET")
	req := httptest.NewRequest("GET", "/accounts/acc1/predict?days=90", nil)
	req = mux.SetURLVars(req, map[string]string{"accountId": "acc1"})
	req = req.WithContext(context.WithValue(req.Context(), "userID", "user1"))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "acc1", resp["account_id"])
}
