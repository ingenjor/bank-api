package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort        string
	DBUrl             string
	JWTSecret         string
	SMTPHost          string
	SMTPPort          int
	SMTPUser          string
	SMTPPass          string
	PGPPublicKeyPath  string
	PGPPrivateKeyPath string
	PGPPassphrase     string
	HMACSecret        string
}

func Load() (*Config, error) {
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	return &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		DBUrl:             os.Getenv("DATABASE_URL"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		SMTPHost:          os.Getenv("SMTP_HOST"),
		SMTPPort:          smtpPort,
		SMTPUser:          os.Getenv("SMTP_USER"),
		SMTPPass:          os.Getenv("SMTP_PASSWORD"),
		PGPPublicKeyPath:  os.Getenv("PGP_PUBLIC_KEY_PATH"),
		PGPPrivateKeyPath: os.Getenv("PGP_PRIVATE_KEY_PATH"),
		PGPPassphrase:     os.Getenv("PGP_PASSPHRASE"),
		HMACSecret:        os.Getenv("HMAC_SECRET"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
