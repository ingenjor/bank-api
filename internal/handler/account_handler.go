package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"bank-api/internal/models"
)

type AccountService interface {
	Create(ctx context.Context, userID string) (*models.Account, error)
	GetUserAccounts(ctx context.Context, userID string) ([]*models.Account, error)
	Deposit(ctx context.Context, accountID, userID string, amount decimal.Decimal) error
	Withdraw(ctx context.Context, accountID, userID string, amount decimal.Decimal) error
}

type AccountHandler struct {
	svc       AccountService
	logger    *logrus.Logger
	validator *validator.Validate
}

func NewAccountHandler(s AccountService, l *logrus.Logger) *AccountHandler {
	return &AccountHandler{svc: s, logger: l, validator: validator.New()}
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	acc, err := h.svc.Create(r.Context(), userID)
	if err != nil {
		respond(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusCreated, acc)
}

func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	accounts, err := h.svc.GetUserAccounts(r.Context(), userID)
	if err != nil {
		respond(w, http.StatusInternalServerError, err.Error())
		return
	}
	json.NewEncoder(w).Encode(accounts)
}

func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	vars := mux.Vars(r)
	accountID := vars["id"]
	var req models.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		respond(w, http.StatusBadRequest, err.Error())
		return
	}
	amount := decimal.NewFromFloat(req.Amount)
	if err := h.svc.Deposit(r.Context(), accountID, userID, amount); err != nil {
		respond(w, http.StatusForbidden, err.Error())
		return
	}
	respond(w, http.StatusOK, "deposit successful")
}

func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	vars := mux.Vars(r)
	accountID := vars["id"]
	var req models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		respond(w, http.StatusBadRequest, err.Error())
		return
	}
	amount := decimal.NewFromFloat(req.Amount)
	if err := h.svc.Withdraw(r.Context(), accountID, userID, amount); err != nil {
		respond(w, http.StatusForbidden, err.Error())
		return
	}
	respond(w, http.StatusOK, "withdrawal successful")
}
