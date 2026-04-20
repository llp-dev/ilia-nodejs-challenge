package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port              string
	Release           bool
	DSN               string
	JWTSecret         string
	JWTInternalSecret string
	WalletURL         string
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("USERS_PORT")
	if port == "" {
		return nil, fmt.Errorf("USERS_PORT environment variable is required")
	}

	dsn := os.Getenv("USERS_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("USERS_DSN environment variable is required")
	}

	jwtSecret := os.Getenv("USERS_JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("USERS_JWT_SECRET environment variable is required")
	}

	jwtInternalSecret := os.Getenv("USERS_JWT_INTERNAL_SECRET")
	if jwtInternalSecret == "" {
		return nil, fmt.Errorf("USERS_JWT_INTERNAL_SECRET environment variable is required")
	}

	walletURL := os.Getenv("USERS_WALLET_URL")
	if walletURL == "" {
		return nil, fmt.Errorf("USERS_WALLET_URL environment variable is required")
	}

	release := os.Getenv("USERS_RELEASE") == "true"

	return &Config{
		Port:              port,
		Release:           release,
		DSN:               dsn,
		JWTSecret:         jwtSecret,
		JWTInternalSecret: jwtInternalSecret,
		WalletURL:         walletURL,
	}, nil
}
