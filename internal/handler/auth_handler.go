package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"

	"bank-api/internal/models"
)

type AuthService interface {
	Register(ctx context.Context, req models.RegisterRequest) error
	Login(ctx context.Context, req models.LoginRequest) (string, error)
}

type AuthHandler struct {
	svc       AuthService
	logger    *logrus.Logger
	validator *validator.Validate
}

func NewAuthHandler(s AuthService, l *logrus.Logger) *AuthHandler {
	return &AuthHandler{svc: s, logger: l, validator: validator.New()}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		respond(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.Register(r.Context(), req); err != nil {
		h.logger.WithError(err).Warn("registration failed")
		respond(w, http.StatusConflict, err.Error())
		return
	}
	respond(w, http.StatusCreated, "registered successfully")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, "invalid request")
		return
	}
	if err := h.validator.Struct(req); err != nil {
		respond(w, http.StatusBadRequest, err.Error())
		return
	}
	token, err := h.svc.Login(r.Context(), req)
	if err != nil {
		respond(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	json.NewEncoder(w).Encode(models.LoginResponse{Token: token})
}
