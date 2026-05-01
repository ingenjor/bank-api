package handler

import (
	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type CreditService interface {
	Apply(ctx context.Context, userID string, req models.CreditApplicationRequest) (*models.Credit, error)
	GetSchedule(ctx context.Context, creditID, userID string) ([]models.PaymentScheduleItem, error)
}

type CreditHandler struct {
	svc       CreditService
	logger    *logrus.Logger
	validator *validator.Validate
}

func NewCreditHandler(s CreditService, l *logrus.Logger) *CreditHandler {
	return &CreditHandler{svc: s, logger: l, validator: validator.New()}
}

func (h *CreditHandler) Apply(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	var req models.CreditApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		respond(w, http.StatusBadRequest, err.Error())
		return
	}
	credit, err := h.svc.Apply(r.Context(), userID, req)
	if err != nil {
		h.logger.WithError(err).Error("credit application failed")
		respond(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusCreated, credit)
}

func (h *CreditHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	creditID := vars["creditId"]
	schedule, err := h.svc.GetSchedule(r.Context(), creditID, userID)
	if err != nil {
		respond(w, http.StatusForbidden, err.Error())
		return
	}
	json.NewEncoder(w).Encode(schedule)
}
