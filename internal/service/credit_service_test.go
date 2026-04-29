package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"bank-api/internal/integration"
	"bank-api/internal/models"
	"bank-api/internal/service"
)

type mockCreditRepo struct {
	mock.Mock
}

func (m *mockCreditRepo) Create(ctx context.Context, c *models.Credit) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}
func (m *mockCreditRepo) GetByID(ctx context.Context, id string) (*models.Credit, error) {
	return nil, nil
}
func (m *mockCreditRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Credit, error) {
	return nil, nil
}
func (m *mockCreditRepo) AddSchedule(ctx context.Context, creditID string, schedule []models.PaymentScheduleItem) error {
	args := m.Called(ctx, creditID, schedule)
	return args.Error(0)
}
func (m *mockCreditRepo) GetSchedule(ctx context.Context, creditID string) ([]models.PaymentScheduleItem, error) {
	return nil, nil
}
func (m *mockCreditRepo) GetOverduePayments(ctx context.Context) ([]*models.Credit, error) {
	return nil, nil
}
func (m *mockCreditRepo) Update(ctx context.Context, c *models.Credit) error { return nil }
func (m *mockCreditRepo) MarkPaymentAsPaid(ctx context.Context, creditID string, dueDate time.Time) error {
	return nil
}
func (m *mockCreditRepo) HasPenaltyBeenApplied(ctx context.Context, creditID string, dueDate time.Time) (bool, error) {
	return false, nil
}
func (m *mockCreditRepo) ApplyPenalty(ctx context.Context, creditID string, dueDate time.Time) error {
	return nil
}

type mockAccRepoForCredit struct {
	mock.Mock
}

func (m *mockAccRepoForCredit) GetByUserID(ctx context.Context, userID string) ([]*models.Account, error) {
	return nil, nil
}
func (m *mockAccRepoForCredit) UpdateBalance(ctx context.Context, id string, newBalance decimal.Decimal) error {
	return nil
}
func (m *mockAccRepoForCredit) Create(ctx context.Context, a *models.Account) error { return nil }
func (m *mockAccRepoForCredit) GetByID(ctx context.Context, id string) (*models.Account, error) {
	return nil, nil
}
func (m *mockAccRepoForCredit) TransferTx(ctx context.Context, fromID, toID string, amount decimal.Decimal) error {
	return nil
}

type mockTxRepoForCredit struct {
	mock.Mock
}

func (m *mockTxRepoForCredit) RecordCreditTransaction(ctx context.Context, fromAcc, toAcc string, amount decimal.Decimal, txType string, creditID string) error {
	return nil
}
func (m *mockTxRepoForCredit) Record(ctx context.Context, fromAcc, toAcc string, amount decimal.Decimal, txType string) error {
	return nil
}
func (m *mockTxRepoForCredit) GetMonthlySummary(ctx context.Context, accountID string, monthStart, monthEnd time.Time) (decimal.Decimal, decimal.Decimal, error) {
	return decimal.Zero, decimal.Zero, nil
}
func (m *mockTxRepoForCredit) GetUpcomingCreditPayments(ctx context.Context, userID string, from, to time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}
func (m *mockTxRepoForCredit) GetTotalPenaltiesForUser(ctx context.Context, userID string, start, end time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func TestCreditService_Apply(t *testing.T) {
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
	defer server.Close()

	integration.CBRServiceURL = server.URL
	defer func() { integration.CBRServiceURL = "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx" }()

	creditRepo := new(mockCreditRepo)
	accRepo := new(mockAccRepoForCredit)
	txRepo := new(mockTxRepoForCredit)

	cbrClient := integration.NewCBRClient()
	svc := service.NewCreditService(creditRepo, accRepo, txRepo, cbrClient)

	creditRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Credit")).Return(nil)
	creditRepo.On("AddSchedule", mock.Anything, mock.Anything, mock.AnythingOfType("[]models.PaymentScheduleItem")).Return(nil)

	credit, err := svc.Apply(context.Background(), "user1", models.CreditApplicationRequest{Amount: 100000, TermMonths: 12})
	assert.NoError(t, err)
	assert.Equal(t, "active", credit.Status)
}
