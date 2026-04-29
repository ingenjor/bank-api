package models

type Card struct {
	ID              string
	AccountID       string
	EncryptedNumber []byte `json:"-"`
	HMACNumber      string `json:"-"`
	EncryptedExpiry []byte `json:"-"`
	CVVHash         string `json:"-"`
	Status          string `json:"status"`
	LastFour        string `json:"-"`
}

type CardResponse struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	Number    string `json:"number"`
	Expiry    string `json:"expiry"`
}

type CreateCardRequest struct {
	AccountID string `json:"account_id" validate:"required,uuid"`
}

type PaymentRequest struct {
	CardID string  `json:"card_id" validate:"required,uuid"`
	Amount float64 `json:"amount"  validate:"required,gt=0"`
}
