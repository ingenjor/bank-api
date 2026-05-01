package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"bank-api/internal/middleware"
	"bank-api/internal/service"
)

type AnalyticsService interface {
	GetAnalytics(ctx context.Context, userID string) (*service.AnalyticsData, error)
	PredictBalance(ctx context.Context, accountID, userID string, days int) (decimal.Decimal, error)
}

type AnalyticsHandler struct {
	svc    AnalyticsService
	logger *logrus.Logger
}

func NewAnalyticsHandler(s AnalyticsService, l *logrus.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{svc: s, logger: l}
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	data, err := h.svc.GetAnalytics(r.Context(), userID)
	if err != nil {
		respond(w, http.StatusInternalServerError, err.Error())
		return
	}
	json.NewEncoder(w).Encode(data)
}

func (h *AnalyticsHandler) PredictBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	accountID := vars["accountId"]
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		respond(w, http.StatusBadRequest, "days must be between 1 and 365")
		return
	}
	prediction, err := h.svc.PredictBalance(r.Context(), accountID, userID, days)
	if err != nil {
		respond(w, http.StatusForbidden, err.Error())
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"account_id":        accountID,
		"predicted_balance": prediction,
		"days":              days,
	})
}
