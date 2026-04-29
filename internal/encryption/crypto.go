package encryption

import (
	"bytes"
	"fmt"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"
	"golang.org/x/crypto/bcrypt"
)

type CryptoService struct {
	publicKey  openpgp.EntityList
	privateKey *openpgp.Entity
}

func NewCryptoService(pubKeyPath, privKeyPath, passphrase string) (*CryptoService, error) {
	svc := &CryptoService{}

	pubData, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading public key: %w", err)
	}
	entities, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(pubData))
	if err != nil {
		return nil, fmt.Errorf("parsing public key: %w", err)
	}
	svc.publicKey = entities

	privData, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}
	privEntities, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(privData))
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}
	if len(privEntities) == 0 {
		return nil, fmt.Errorf("no private key found")
	}
	priv := privEntities[0]
	if priv.PrivateKey != nil && priv.PrivateKey.Encrypted {
		if err := priv.PrivateKey.Decrypt([]byte(passphrase)); err != nil {
			return nil, fmt.Errorf("failed to decrypt private key: %w", err)
		}
	}
	svc.privateKey = priv
	return svc, nil
}

func (s *CryptoService) Encrypt(plaintext []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, s.publicKey, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(plaintext); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *CryptoService) Decrypt(ciphertext []byte) ([]byte, error) {
	md, err := openpgp.ReadMessage(bytes.NewReader(ciphertext), openpgp.EntityList{s.privateKey}, nil, nil)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(md.UnverifiedBody); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func HashCVV(cvv string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	return string(bytes), err
}
