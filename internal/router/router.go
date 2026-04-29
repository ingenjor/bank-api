package router

import (
	"github.com/gorilla/mux"

	"bank-api/internal/config"
	"bank-api/internal/handler"
	"bank-api/internal/middleware"
)

func Setup(cfg *config.Config,
	authH *handler.AuthHandler,
	accH *handler.AccountHandler,
	cardH *handler.CardHandler,
	transH *handler.TransferHandler,
	credH *handler.CreditHandler,
	analH *handler.AnalyticsHandler,
) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/register", authH.Register).Methods("POST")
	r.HandleFunc("/login", authH.Login).Methods("POST")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	api.HandleFunc("/accounts", accH.Create).Methods("POST")
	api.HandleFunc("/accounts", accH.GetAccounts).Methods("GET")
	api.HandleFunc("/accounts/{id}/deposit", accH.Deposit).Methods("POST")
	api.HandleFunc("/accounts/{id}/withdraw", accH.Withdraw).Methods("POST")

	api.HandleFunc("/cards", cardH.Issue).Methods("POST")
	api.HandleFunc("/cards", cardH.GetCards).Methods("GET")
	api.HandleFunc("/cards/{id}", cardH.GetCard).Methods("GET")
	api.HandleFunc("/cards/payment", cardH.Payment).Methods("POST")

	api.HandleFunc("/transfer", transH.Transfer).Methods("POST")

	api.HandleFunc("/credits", credH.Apply).Methods("POST")
	api.HandleFunc("/credits/{creditId}/schedule", credH.GetSchedule).Methods("GET")

	api.HandleFunc("/analytics", analH.GetAnalytics).Methods("GET")
	api.HandleFunc("/accounts/{accountId}/predict", analH.PredictBalance).Methods("GET")

	return r
}
