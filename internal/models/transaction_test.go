package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"bank-api/internal/models"
)

func TestTransactionJSON(t *testing.T) {
	tx := models.Transaction{
		ID:            "tx1",
		FromAccountID: "from",
		ToAccountID:   "to",
		Amount:        decimal.NewFromFloat(150.75),
		Type:          "transfer",
		Timestamp:     time.Now(),
	}
	data, err := json.Marshal(tx)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "from_account_id")
	assert.Contains(t, string(data), "150.75")
}
