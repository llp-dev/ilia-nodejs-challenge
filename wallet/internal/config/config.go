package config

import (
	"os"
)

type Config struct {
	Port    string
	Release bool
}

func LoadConfig() *Config {
	port := os.Getenv("WALLET_PORT")
	if port == "" {
		port = "3001"
	}

	release := os.Getenv("WALLET_RELEASE") == "true"

	return &Config{
		Port:    port,
		Release: release,
	}
}
