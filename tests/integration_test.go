//go:build integration
// +build integration

package tests

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bank-api/internal/encryption"
	"bank-api/internal/integration"
	"bank-api/internal/models"
	"bank-api/internal/repository"
	"bank-api/internal/service"
)

func getTestDB(t *testing.T) *sql.DB {
	_ = godotenv.Load("../.env")
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	err = db.Ping()
	require.NoError(t, err)
	return db
}

func cleanDB(t *testing.T, db *sql.DB) {
	tables := []string{
		"payment_schedules",
		"transactions",
		"cards",
		"credits",
		"accounts",
		"users",
	}
	for _, table := range tables {
		_, err := db.Exec("DELETE FROM " + table)
		if err != nil {
			t.Logf("Warning: could not clean table %s: %v", table, err)
		}
	}
}

type dummyEmailSender struct{}

func (d *dummyEmailSender) Send(to, subject, body string) error {
	return nil
}

func startMockCBRSrv(t *testing.T) *httptest.Server {
	t.Helper()
	xmlResponse := `<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
	  <soap:Body>
	    <KeyRateResponse xmlns="http://web.cbr.ru/">
	      <KeyRateResult>
	        <xsd:schema>schema</xsd:schema>xml<diffgr:diffgram xmlns:diffgr="urn:schemas-microsoft-com:xml-diffgram-v1">
	          <KeyRate xmlns="">
	            <KR>
	              <Rate>7.50</Rate>
	            </KR>
	          </KeyRate>
	        </diffgr:diffgram>
	      </KeyRateResult>
	    </KeyRateResponse>
	  </soap:Body>
	</soap:Envelope>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.Write([]byte(xmlResponse))
	}))
	return server
}

func TestFullFlow_RegistrationToTransfer(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	cleanDB(t, db)

	userRepo := repository.NewUserRepo(db)
	accountRepo := repository.NewAccountRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)

	authService := service.NewAuthService(userRepo, "test-secret")
	accountService := service.NewAccountService(accountRepo, transactionRepo)
	transactionService := service.NewTransactionService(transactionRepo, accountRepo, &dummyEmailSender{}, userRepo)

	ctx := context.Background()

	regReq := models.RegisterRequest{Username: "alice", Email: "alice@example.com", Password: "password123"}
	err := authService.Register(ctx, regReq)
	require.NoError(t, err)

	token, err := authService.Login(ctx, models.LoginRequest{Email: "alice@example.com", Password: "password123"})
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	alice, _ := userRepo.GetByEmail(ctx, "alice@example.com")
	accAlice, err := accountService.Create(ctx, alice.ID)
	require.NoError(t, err)
	err = accountService.Deposit(ctx, accAlice.ID, alice.ID, decimal.NewFromInt(10000))
	require.NoError(t, err)

	_ = authService.Register(ctx, models.RegisterRequest{Username: "bob", Email: "bob@example.com", Password: "password123"})
	bob, _ := userRepo.GetByEmail(ctx, "bob@example.com")
	accBob, err := accountService.Create(ctx, bob.ID)
	require.NoError(t, err)

	err = transactionService.Transfer(ctx, accAlice.ID, accBob.ID, alice.ID, decimal.NewFromInt(2500))
	require.NoError(t, err)

	aliceAfter, _ := accountRepo.GetByID(ctx, accAlice.ID)
	bobAfter, _ := accountRepo.GetByID(ctx, accBob.ID)
	assert.True(t, decimal.NewFromInt(7500).Equal(aliceAfter.Balance), "expected 7500, got %s", aliceAfter.Balance)
	assert.True(t, decimal.NewFromInt(2500).Equal(bobAfter.Balance), "expected 2500, got %s", bobAfter.Balance)
}

func TestCardFlow_IssueAndPayment(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	cleanDB(t, db)

	userRepo := repository.NewUserRepo(db)
	accountRepo := repository.NewAccountRepo(db)
	cardRepo := repository.NewCardRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)

	cryptoService, err := encryption.NewCryptoService("../keys/public.asc", "../keys/private.asc", "")
	require.NoError(t, err, "Cannot load PGP keys for integration test")

	authService := service.NewAuthService(userRepo, "test-secret")
	accountService := service.NewAccountService(accountRepo, transactionRepo)
	cardService := service.NewCardService(cardRepo, accountRepo, cryptoService, "hmac-key", transactionRepo, &dummyEmailSender{}, userRepo)

	ctx := context.Background()

	err = authService.Register(ctx, models.RegisterRequest{Username: "carduser", Email: "card@example.com", Password: "password123"})
	require.NoError(t, err)
	user, _ := userRepo.GetByEmail(ctx, "card@example.com")
	acc, _ := accountService.Create(ctx, user.ID)
	_ = accountService.Deposit(ctx, acc.ID, user.ID, decimal.NewFromInt(50000))

	cardResp, err := cardService.Issue(ctx, user.ID, acc.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, cardResp.Number)
	assert.Len(t, cardResp.Number, 16)

	err = cardService.Payment(ctx, user.ID, models.PaymentRequest{CardID: cardResp.ID, Amount: 15000})
	require.NoError(t, err)

	accAfter, _ := accountRepo.GetByID(ctx, acc.ID)
	assert.True(t, decimal.NewFromInt(35000).Equal(accAfter.Balance), "expected 35000, got %s", accAfter.Balance)
}

func TestCreditFlow_AutoPaymentAndPenalty(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	cleanDB(t, db)

	userRepo := repository.NewUserRepo(db)
	accountRepo := repository.NewAccountRepo(db)
	creditRepo := repository.NewCreditRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)

	server := startMockCBRSrv(t)
	defer server.Close()
	integration.CBRServiceURL = server.URL
	defer func() { integration.CBRServiceURL = "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx" }()

	cbrClient := integration.NewCBRClient()
	authService := service.NewAuthService(userRepo, "test-secret")
	accountService := service.NewAccountService(accountRepo, transactionRepo)
	creditService := service.NewCreditService(creditRepo, accountRepo, transactionRepo, cbrClient)

	ctx := context.Background()

	err := authService.Register(ctx, models.RegisterRequest{Username: "creditauto", Email: "creditauto@example.com", Password: "password123"})
	require.NoError(t, err)
	user, _ := userRepo.GetByEmail(ctx, "creditauto@example.com")

	acc, _ := accountService.Create(ctx, user.ID)
	_ = accountService.Deposit(ctx, acc.ID, user.ID, decimal.NewFromInt(20000))

	credit, err := creditService.Apply(ctx, user.ID, models.CreditApplicationRequest{Amount: 12000, TermMonths: 3})
	require.NoError(t, err)

	monthlyPayment := credit.MonthlyPayment.Round(2)

	_, err = db.Exec("UPDATE credits SET next_payment_date = '1970-01-01' WHERE id = $1", credit.ID)
	require.NoError(t, err)

	_, err = creditService.ProcessOverdue(ctx)
	require.NoError(t, err)

	accAfterPayment, _ := accountRepo.GetByID(ctx, acc.ID)
	expectedBalance := decimal.NewFromInt(20000).Sub(monthlyPayment).Round(2)
	assert.True(t, expectedBalance.Equal(accAfterPayment.Balance.Round(2)),
		"balance should be %s, got %s", expectedBalance, accAfterPayment.Balance.Round(2))

	updatedCredit, _ := creditRepo.GetByID(ctx, credit.ID)
	expectedRemaining := decimal.NewFromInt(12000).Sub(monthlyPayment).Round(2)
	assert.True(t, expectedRemaining.Equal(updatedCredit.Remaining.Round(2)),
		"remaining should be %s, got %s", expectedRemaining, updatedCredit.Remaining.Round(2))

	currentBalance := accAfterPayment.Balance
	withdrawAmount := currentBalance.Sub(decimal.NewFromInt(100)).Round(2)
	err = accountService.Withdraw(ctx, acc.ID, user.ID, withdrawAmount)
	require.NoError(t, err)

	_, err = db.Exec("UPDATE credits SET next_payment_date = '1970-01-01' WHERE id = $1", credit.ID)
	require.NoError(t, err)

	_, err = creditService.ProcessOverdue(ctx)
	require.NoError(t, err)

	creditAfterPenalty, _ := creditRepo.GetByID(ctx, credit.ID)
	penalty := monthlyPayment.Mul(decimal.NewFromFloat(0.1)).Round(2)
	expectedRemainingAfterPenalty := updatedCredit.Remaining.Add(penalty).Round(2)
	assert.True(t, creditAfterPenalty.Remaining.Round(2).Equal(expectedRemainingAfterPenalty),
		"remaining should be %s, got %s", expectedRemainingAfterPenalty, creditAfterPenalty.Remaining.Round(2))

	now := time.Now()
	penalties, err := transactionRepo.GetTotalPenaltiesForUser(ctx, user.ID, now.AddDate(0, 0, -1), now.AddDate(0, 0, 1))
	require.NoError(t, err)
	assert.True(t, penalties.GreaterThan(decimal.Zero), "penalty transaction should exist")
	assert.Equal(t, "active", creditAfterPenalty.Status)
}

func TestCreditFlow_FullRepayment(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	cleanDB(t, db)

	userRepo := repository.NewUserRepo(db)
	accountRepo := repository.NewAccountRepo(db)
	creditRepo := repository.NewCreditRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)

	server := startMockCBRSrv(t)
	defer server.Close()
	integration.CBRServiceURL = server.URL
	defer func() { integration.CBRServiceURL = "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx" }()

	cbrClient := integration.NewCBRClient()
	authService := service.NewAuthService(userRepo, "test-secret")
	accountService := service.NewAccountService(accountRepo, transactionRepo)
	creditService := service.NewCreditService(creditRepo, accountRepo, transactionRepo, cbrClient)

	ctx := context.Background()

	err := authService.Register(ctx, models.RegisterRequest{Username: "fullpayer", Email: "fullpayer@example.com", Password: "password123"})
	require.NoError(t, err)
	user, _ := userRepo.GetByEmail(ctx, "fullpayer@example.com")

	acc, _ := accountService.Create(ctx, user.ID)
	_ = accountService.Deposit(ctx, acc.ID, user.ID, decimal.NewFromInt(50000))

	credit, err := creditService.Apply(ctx, user.ID, models.CreditApplicationRequest{Amount: 36000, TermMonths: 12})
	require.NoError(t, err)
	require.Equal(t, "active", credit.Status)

	for i := 0; i < 15; i++ {
		_, err = db.Exec("UPDATE credits SET next_payment_date = '1970-01-01' WHERE id = $1", credit.ID)
		require.NoError(t, err)

		_, err = creditService.ProcessOverdue(ctx)
		require.NoError(t, err)

		credit, err = creditRepo.GetByID(ctx, credit.ID)
		require.NoError(t, err)
		if credit.Status == "paid" {
			break
		}
	}

	assert.Equal(t, "paid", credit.Status, "credit should be fully paid after enough payments")
	assert.True(t, credit.Remaining.LessThanOrEqual(decimal.Zero), "remaining should be zero or negative, got %s", credit.Remaining)
}

func TestAnalyticsFlow_IncomeExpenseAndPrediction(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	cleanDB(t, db)

	userRepo := repository.NewUserRepo(db)
	accountRepo := repository.NewAccountRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)
	creditRepo := repository.NewCreditRepo(db)

	authService := service.NewAuthService(userRepo, "test-secret")
	accountService := service.NewAccountService(accountRepo, transactionRepo)
	analyticsService := service.NewAnalyticsService(transactionRepo, accountRepo, creditRepo)

	ctx := context.Background()

	err := authService.Register(ctx, models.RegisterRequest{Username: "analuser", Email: "anal@example.com", Password: "password123"})
	require.NoError(t, err)
	user, _ := userRepo.GetByEmail(ctx, "anal@example.com")

	acc1, _ := accountService.Create(ctx, user.ID)
	acc2, _ := accountService.Create(ctx, user.ID)

	_ = accountService.Deposit(ctx, acc1.ID, user.ID, decimal.NewFromInt(20000))
	_ = accountService.Deposit(ctx, acc2.ID, user.ID, decimal.NewFromInt(10000))

	_ = accountService.Withdraw(ctx, acc1.ID, user.ID, decimal.NewFromInt(5000))
	_ = service.NewTransactionService(transactionRepo, accountRepo, &dummyEmailSender{}, userRepo).
		Transfer(ctx, acc1.ID, acc2.ID, user.ID, decimal.NewFromInt(2000))

	data, err := analyticsService.GetAnalytics(ctx, user.ID)
	require.NoError(t, err)

	assert.True(t, decimal.NewFromInt(32000).Equal(data.MonthlyIncome), "expected 32000, got %s", data.MonthlyIncome)
	assert.True(t, decimal.NewFromInt(7000).Equal(data.MonthlyExpense), "expected 7000, got %s", data.MonthlyExpense)

	pred, err := analyticsService.PredictBalance(ctx, acc1.ID, user.ID, 30)
	require.NoError(t, err)
	assert.True(t, decimal.NewFromInt(13000).Equal(pred), "expected 13000, got %s", pred)
}
