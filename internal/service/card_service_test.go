package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"bank-api/internal/encryption"
	"bank-api/internal/models"
	"bank-api/internal/service"
)

type mockCardRepo struct{ mock.Mock }
type mockAccRepoForCard struct{ mock.Mock }
type mockTxRepoForCard struct{ mock.Mock }
type mockUserRepoForCard struct{ mock.Mock }

func (m *mockCardRepo) Create(ctx context.Context, card *models.Card) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}
func (m *mockCardRepo) GetByID(ctx context.Context, id string) (*models.Card, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Card), args.Error(1)
}
func (m *mockCardRepo) GetByAccountID(ctx context.Context, accountID string) ([]*models.Card, error) {
	return nil, nil
}
func (m *mockCardRepo) GetCardsByUserID(ctx context.Context, userID string) ([]*models.Card, error) {
	return nil, nil
}

func (m *mockAccRepoForCard) GetByID(ctx context.Context, id string) (*models.Account, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Account), args.Error(1)
}
func (m *mockAccRepoForCard) Create(ctx context.Context, a *models.Account) error { return nil }
func (m *mockAccRepoForCard) GetByUserID(ctx context.Context, userID string) ([]*models.Account, error) {
	return nil, nil
}
func (m *mockAccRepoForCard) UpdateBalance(ctx context.Context, id string, newBalance decimal.Decimal) error {
	args := m.Called(ctx, id, newBalance)
	return args.Error(0)
}
func (m *mockAccRepoForCard) TransferTx(ctx context.Context, fromID, toID string, amount decimal.Decimal) error {
	return nil
}

func (m *mockTxRepoForCard) Record(ctx context.Context, fromAcc, toAcc string, amount decimal.Decimal, txType string) error {
	args := m.Called(ctx, fromAcc, toAcc, amount, txType)
	return args.Error(0)
}
func (m *mockTxRepoForCard) RecordCreditTransaction(ctx context.Context, fromAcc, toAcc string, amount decimal.Decimal, txType string, creditID string) error {
	return nil
}
func (m *mockTxRepoForCard) GetMonthlySummary(ctx context.Context, accountID string, monthStart, monthEnd time.Time) (decimal.Decimal, decimal.Decimal, error) {
	return decimal.Zero, decimal.Zero, nil
}
func (m *mockTxRepoForCard) GetUpcomingCreditPayments(ctx context.Context, userID string, from, to time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}
func (m *mockTxRepoForCard) GetTotalPenaltiesForUser(ctx context.Context, userID string, start, end time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func (m *mockUserRepoForCard) GetByID(ctx context.Context, id string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepoForCard) Create(ctx context.Context, u *models.User) error { return nil }
func (m *mockUserRepoForCard) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepoForCard) IsUnique(ctx context.Context, email, username string) (bool, error) {
	return true, nil
}

// createTempPGPKeys генерирует временные armored-ключи для тестов.
func createTempPGPKeys(t *testing.T) (pubPath, privPath string) {
	t.Helper()
	dir := t.TempDir()
	pubPath = filepath.Join(dir, "public.asc")
	privPath = filepath.Join(dir, "private.asc")

	entity, err := openpgp.NewEntity("test", "test", "test@test.com", nil)
	require.NoError(t, err)

	// Публичный ключ (armored)
	pubFile, err := os.Create(pubPath)
	require.NoError(t, err)
	defer pubFile.Close()
	pubWriter, err := armor.Encode(pubFile, openpgp.PublicKeyType, nil)
	require.NoError(t, err)
	err = entity.Serialize(pubWriter)
	require.NoError(t, err)
	pubWriter.Close()

	// Приватный ключ (armored)
	privFile, err := os.Create(privPath)
	require.NoError(t, err)
	defer privFile.Close()
	privWriter, err := armor.Encode(privFile, openpgp.PrivateKeyType, nil)
	require.NoError(t, err)
	err = entity.SerializePrivate(privWriter, nil)
	require.NoError(t, err)
	privWriter.Close()

	return
}

func TestCardService_Issue_Success(t *testing.T) {
	cardRepo := new(mockCardRepo)
	accRepo := new(mockAccRepoForCard)
	txRepo := new(mockTxRepoForCard)
	userRepo := new(mockUserRepoForCard)

	pubPath, privPath := createTempPGPKeys(t)
	cryptoService, err := encryption.NewCryptoService(pubPath, privPath, "")
	require.NoError(t, err)

	svc := service.NewCardService(cardRepo, accRepo, cryptoService, "hmac-secret", txRepo, nil, userRepo)

	acc := &models.Account{ID: "acc1", UserID: "user1", Balance: decimal.NewFromInt(1000)}
	accRepo.On("GetByID", mock.Anything, "acc1").Return(acc, nil)
	cardRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Card")).Return(nil)

	cardResp, err := svc.Issue(context.Background(), "user1", "acc1")
	require.NoError(t, err)
	assert.NotEmpty(t, cardResp.Number)
	assert.Len(t, cardResp.Number, 16)
}

func TestCardService_Payment_InsufficientFunds(t *testing.T) {
	cardRepo := new(mockCardRepo)
	accRepo := new(mockAccRepoForCard)
	txRepo := new(mockTxRepoForCard)
	userRepo := new(mockUserRepoForCard)

	pubPath, privPath := createTempPGPKeys(t)
	cryptoService, err := encryption.NewCryptoService(pubPath, privPath, "")
	require.NoError(t, err)

	svc := service.NewCardService(cardRepo, accRepo, cryptoService, "hmac", txRepo, nil, userRepo)

	card := &models.Card{ID: "card1", AccountID: "acc1", LastFour: "1234", EncryptedNumber: []byte("enc"), HMACNumber: "hmac", EncryptedExpiry: []byte("enc"), CVVHash: "hash"}
	acc := &models.Account{ID: "acc1", UserID: "user1", Balance: decimal.NewFromInt(50)}

	cardRepo.On("GetByID", mock.Anything, "card1").Return(card, nil)
	accRepo.On("GetByID", mock.Anything, "acc1").Return(acc, nil)

	req := models.PaymentRequest{CardID: "card1", Amount: 100}
	err = svc.Payment(context.Background(), "user1", req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")
}
