package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port      string
	Release   bool
	DSN       string
	JWTSecret string
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("WALLET_PORT")
	if port == "" {
		return nil, fmt.Errorf("WALLET_PORT environment variable is required")
	}

	dsn := os.Getenv("WALLET_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("WALLET_DSN environment variable is required")
	}

	jwtSecret := os.Getenv("WALLET_JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("WALLET_JWT_SECRET environment variable is required")
	}

	release := os.Getenv("WALLET_RELEASE") == "true"

	return &Config{
		Port:      port,
		Release:   release,
		DSN:       dsn,
		JWTSecret: jwtSecret,
	}, nil
}
