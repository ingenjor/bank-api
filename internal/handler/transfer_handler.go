package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type TransactionService interface {
	Transfer(ctx context.Context, fromID, toID, userID string, amount decimal.Decimal) error
}

type TransferHandler struct {
	svc       TransactionService
	logger    *logrus.Logger
	validator *validator.Validate
}

func NewTransferHandler(s TransactionService, l *logrus.Logger) *TransferHandler {
	return &TransferHandler{svc: s, logger: l, validator: validator.New()}
}

func (h *TransferHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	var req struct {
		From   string  `json:"from"   validate:"required,uuid"`
		To     string  `json:"to"     validate:"required,uuid"`
		Amount float64 `json:"amount" validate:"required,gt=0"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		respond(w, http.StatusBadRequest, err.Error())
		return
	}
	amount := decimal.NewFromFloat(req.Amount)
	if err := h.svc.Transfer(r.Context(), req.From, req.To, userID, amount); err != nil {
		respond(w, http.StatusForbidden, err.Error())
		return
	}
	respond(w, http.StatusOK, "transfer completed")
}
