package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"bank-api/internal/encryption"
	"bank-api/internal/integration"
	"bank-api/internal/models"
)

type CardRepository interface {
	Create(ctx context.Context, card *models.Card) error
	GetByID(ctx context.Context, id string) (*models.Card, error)
	GetByAccountID(ctx context.Context, accountID string) ([]*models.Card, error)
	GetCardsByUserID(ctx context.Context, userID string) ([]*models.Card, error)
}

type CardService struct {
	cardRepo    CardRepository
	accountRepo AccountRepository
	crypto      *encryption.CryptoService
	hmacSecret  []byte
	txRepo      TransactionRepository
	emailSender integration.EmailSender
	userRepo    UserRepository
}

func NewCardService(cr CardRepository, ar AccountRepository, crypto *encryption.CryptoService, hmacSecret string, txr TransactionRepository, es integration.EmailSender, ur UserRepository) *CardService {
	return &CardService{
		cardRepo:    cr,
		accountRepo: ar,
		crypto:      crypto,
		hmacSecret:  []byte(hmacSecret),
		txRepo:      txr,
		emailSender: es,
		userRepo:    ur,
	}
}

func generateLuhn() string {
	src := rand.New(rand.NewSource(time.Now().UnixNano()))
	digits := make([]int, 16)
	digits[0], digits[1], digits[2], digits[3] = 4, 2, 7, 6
	for i := 4; i < 15; i++ {
		digits[i] = src.Intn(10)
	}
	sum := 0
	for i := 14; i >= 0; i-- {
		d := digits[i]
		if (14-i)%2 == 1 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	digits[15] = (10 - (sum % 10)) % 10
	card := ""
	for _, d := range digits {
		card += fmt.Sprint(d)
	}
	return card
}

func (s *CardService) Issue(ctx context.Context, userID, accountID string) (*models.CardResponse, error) {
	if s.crypto == nil {
		return nil, errors.New("card service is disabled: PGP keys not configured")
	}
	acc, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil || acc.UserID != userID {
		return nil, errors.New("account not found or access denied")
	}

	cardNumber := generateLuhn()
	expiry := time.Now().AddDate(3, 0, 0).Format("01/06")
	cvv := fmt.Sprintf("%03d", rand.Intn(1000))
	lastFour := cardNumber[len(cardNumber)-4:]

	encNumber, err := s.crypto.Encrypt([]byte(cardNumber))
	if err != nil {
		return nil, fmt.Errorf("encrypt number: %w", err)
	}
	encExpiry, err := s.crypto.Encrypt([]byte(expiry))
	if err != nil {
		return nil, fmt.Errorf("encrypt expiry: %w", err)
	}

	h := hmac.New(sha256.New, s.hmacSecret)
	h.Write([]byte(cardNumber))
	hmacNumber := hex.EncodeToString(h.Sum(nil))

	cvvHash, err := encryption.HashCVV(cvv)
	if err != nil {
		return nil, err
	}

	card := &models.Card{
		ID:              uuid.New().String(),
		AccountID:       accountID,
		EncryptedNumber: encNumber,
		HMACNumber:      hmacNumber,
		EncryptedExpiry: encExpiry,
		CVVHash:         cvvHash,
		Status:          "active",
		LastFour:        lastFour,
	}
	if err := s.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}

	return &models.CardResponse{
		ID:        card.ID,
		AccountID: accountID,
		Number:    cardNumber,
		Expiry:    expiry,
	}, nil
}

func (s *CardService) GetByID(ctx context.Context, cardID, userID string) (*models.CardResponse, error) {
	if s.crypto == nil {
		return nil, errors.New("card service is disabled: PGP keys not configured")
	}
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	acc, err := s.accountRepo.GetByID(ctx, card.AccountID)
	if err != nil || acc.UserID != userID {
		return nil, errors.New("access denied")
	}

	plainNumber, err := s.crypto.Decrypt(card.EncryptedNumber)
	if err != nil {
		return nil, errors.New("decryption failed")
	}
	plainExpiry, err := s.crypto.Decrypt(card.EncryptedExpiry)
	if err != nil {
		return nil, errors.New("decryption failed")
	}

	h := hmac.New(sha256.New, s.hmacSecret)
	h.Write(plainNumber)
	expectedHMAC := hex.EncodeToString(h.Sum(nil))
	if expectedHMAC != card.HMACNumber {
		return nil, errors.New("card data integrity check failed")
	}

	return &models.CardResponse{
		ID:        card.ID,
		AccountID: card.AccountID,
		Number:    string(plainNumber),
		Expiry:    string(plainExpiry),
	}, nil
}

func (s *CardService) GetUserCards(ctx context.Context, userID string) ([]*models.CardResponse, error) {
	if s.crypto == nil {
		return nil, errors.New("card service is disabled: PGP keys not configured")
	}
	cards, err := s.cardRepo.GetCardsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	responses := make([]*models.CardResponse, 0, len(cards))
	for _, card := range cards {
		plainNumber, err := s.crypto.Decrypt(card.EncryptedNumber)
		if err != nil {
			continue
		}
		plainExpiry, err := s.crypto.Decrypt(card.EncryptedExpiry)
		if err != nil {
			continue
		}
		h := hmac.New(sha256.New, s.hmacSecret)
		h.Write(plainNumber)
		if hex.EncodeToString(h.Sum(nil)) != card.HMACNumber {
			continue
		}
		responses = append(responses, &models.CardResponse{
			ID:        card.ID,
			AccountID: card.AccountID,
			Number:    string(plainNumber),
			Expiry:    string(plainExpiry),
		})
	}
	return responses, nil
}

func (s *CardService) Payment(ctx context.Context, userID string, req models.PaymentRequest) error {
	if s.crypto == nil {
		return errors.New("card service is disabled: PGP keys not configured")
	}
	card, err := s.cardRepo.GetByID(ctx, req.CardID)
	if err != nil {
		return err
	}
	acc, err := s.accountRepo.GetByID(ctx, card.AccountID)
	if err != nil || acc.UserID != userID {
		return errors.New("access denied")
	}
	amount := decimal.NewFromFloat(req.Amount)
	if acc.Balance.LessThan(amount) {
		return errors.New("insufficient funds")
	}
	newBalance := acc.Balance.Sub(amount)
	if err := s.accountRepo.UpdateBalance(ctx, acc.ID, newBalance); err != nil {
		return err
	}
	if err := s.txRepo.Record(ctx, acc.ID, "", amount, "payment"); err != nil {
		return err
	}
	if s.emailSender != nil && s.userRepo != nil {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err == nil && user != nil {
			body := fmt.Sprintf("Оплата по карте %s на сумму %.2f RUB выполнена.", card.LastFour, amount.InexactFloat64())
			go s.emailSender.Send(user.Email, "Платеж проведен", body)
		}
	}
	return nil
}
