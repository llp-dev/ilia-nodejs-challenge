package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port    string
	Release bool
	DSN     string
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("WALLET_PORT")
	if port == "" {
		port = "3001"
	}

	dsn := os.Getenv("WALLET_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("WALLET_DSN environment variable is required")
	}

	release := os.Getenv("WALLET_RELEASE") == "true"

	return &Config{
		Port:    port,
		Release: release,
		DSN:     dsn,
	}, nil
}
