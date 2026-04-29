package integration_test

import (
	"bank-api/internal/integration"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmailSender(t *testing.T) {
	sender := integration.NewEmailSender("localhost", 25, "user", "pass")
	assert.NotNil(t, sender)
}
