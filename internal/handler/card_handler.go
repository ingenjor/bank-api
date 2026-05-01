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

type CardService interface {
	Issue(ctx context.Context, userID, accountID string) (*models.CardResponse, error)
	GetByID(ctx context.Context, cardID, userID string) (*models.CardResponse, error)
	GetUserCards(ctx context.Context, userID string) ([]*models.CardResponse, error)
	Payment(ctx context.Context, userID string, req models.PaymentRequest) error
}

type CardHandler struct {
	svc       CardService
	logger    *logrus.Logger
	validator *validator.Validate
}

func NewCardHandler(s CardService, l *logrus.Logger) *CardHandler {
	return &CardHandler{svc: s, logger: l, validator: validator.New()}
}

func (h *CardHandler) Issue(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	var req models.CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		respond(w, http.StatusBadRequest, err.Error())
		return
	}
	cardResp, err := h.svc.Issue(r.Context(), userID, req.AccountID)
	if err != nil {
		respond(w, http.StatusForbidden, err.Error())
		return
	}
	respond(w, http.StatusCreated, cardResp)
}

func (h *CardHandler) GetCard(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	vars := mux.Vars(r)
	cardID := vars["id"]
	card, err := h.svc.GetByID(r.Context(), cardID, userID)
	if err != nil {
		respond(w, http.StatusNotFound, err.Error())
		return
	}
	json.NewEncoder(w).Encode(card)
}

func (h *CardHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	cards, err := h.svc.GetUserCards(r.Context(), userID)
	if err != nil {
		respond(w, http.StatusInternalServerError, err.Error())
		return
	}
	json.NewEncoder(w).Encode(cards)
}

func (h *CardHandler) Payment(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		respond(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.Payment(r.Context(), userID, req); err != nil {
		respond(w, http.StatusForbidden, err.Error())
		return
	}
	respond(w, http.StatusOK, "payment successful")
}
