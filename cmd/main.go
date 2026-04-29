package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"bank-api/internal/config"
	"bank-api/internal/encryption"
	"bank-api/internal/handler"
	"bank-api/internal/integration"
	"bank-api/internal/repository"
	"bank-api/internal/router"
	"bank-api/internal/scheduler"
	"bank-api/internal/service"
)

func main() {
	godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		logger.Fatal(err)
	}
	logger.Info("Database connected")

	userRepo := repository.NewUserRepo(db)
	accountRepo := repository.NewAccountRepo(db)
	cardRepo := repository.NewCardRepo(db)
	creditRepo := repository.NewCreditRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)

	emailSender := integration.NewEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	cbrClient := integration.NewCBRClient()

	cryptoService, err := encryption.NewCryptoService(cfg.PGPPublicKeyPath, cfg.PGPPrivateKeyPath, cfg.PGPPassphrase)
	if err != nil {
		logger.Fatalf("Failed to init PGP: %v", err)
	}

	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	accountService := service.NewAccountService(accountRepo, transactionRepo)
	cardService := service.NewCardService(cardRepo, accountRepo, cryptoService, cfg.HMACSecret, transactionRepo, emailSender, userRepo)
	creditService := service.NewCreditService(creditRepo, accountRepo, transactionRepo, cbrClient)
	transactionService := service.NewTransactionService(transactionRepo, accountRepo, emailSender, userRepo)
	analyticsService := service.NewAnalyticsService(transactionRepo, accountRepo, creditRepo)

	authH := handler.NewAuthHandler(authService, logger)
	accountH := handler.NewAccountHandler(accountService, logger)
	cardH := handler.NewCardHandler(cardService, logger)
	transferH := handler.NewTransferHandler(transactionService, logger)
	creditH := handler.NewCreditHandler(creditService, logger)
	analyticsH := handler.NewAnalyticsHandler(analyticsService, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched := scheduler.NewPaymentScheduler(creditService, userRepo, emailSender, 12*time.Hour, logger)
	go sched.Start(ctx)

	r := router.Setup(cfg, authH, accountH, cardH, transferH, creditH, analyticsH)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("Shutting down...")
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	logger.Infof("Server started on :%s", cfg.ServerPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}
	logger.Info("Server stopped")
}
