package encryption_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bank-api/internal/encryption"
)

func TestCryptoService_EncryptDecrypt(t *testing.T) {
	if _, err := os.Stat("../../keys/public.asc"); os.IsNotExist(err) {
		t.Skip("PGP keys not found")
	}
	svc, err := encryption.NewCryptoService("../../keys/public.asc", "../../keys/private.asc", "")
	require.NoError(t, err)
	plaintext := []byte("4111111111111111")
	cipher, err := svc.Encrypt(plaintext)
	require.NoError(t, err)
	decrypted, err := svc.Decrypt(cipher)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestHashPassword(t *testing.T) {
	hash, err := encryption.HashPassword("mypassword")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestHashCVV(t *testing.T) {
	hash, err := encryption.HashCVV("123")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
}
