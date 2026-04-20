package config

import (
	"os"
	"testing"
)

func withRequiredEnvs(t *testing.T) {
	t.Helper()
	t.Setenv("USERS_DSN", "postgres://user:pass@localhost:5432/users")
	t.Setenv("USERS_PORT", "3002")
	t.Setenv("USERS_JWT_SECRET", "ILIACHALLENGE")
	t.Setenv("USERS_JWT_INTERNAL_SECRET", "ILIACHALLENGE_INTERNAL")
	t.Setenv("USERS_WALLET_URL", "http://wallet-api:3001")
}

func TestLoadConfig_Port(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantPort string
		wantErr  bool
	}{
		{name: "uses USERS_PORT when set", envValue: "8080", wantPort: "8080"},
		{name: "errors when USERS_PORT is not set", wantErr: true},
		{name: "accepts non-standard port", envValue: "9999", wantPort: "9999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withRequiredEnvs(t)
			if tt.envValue != "" {
				t.Setenv("USERS_PORT", tt.envValue)
			} else {
				os.Unsetenv("USERS_PORT")
			}

			cfg, err := LoadConfig()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadConfig() returned error: %v", err)
			}
			if cfg.Port != tt.wantPort {
				t.Errorf("Port = %q, want %q", cfg.Port, tt.wantPort)
			}
		})
	}
}

func TestLoadConfig_DSN(t *testing.T) {
	t.Run("errors when USERS_DSN is not set", func(t *testing.T) {
		withRequiredEnvs(t)
		os.Unsetenv("USERS_DSN")
		_, err := LoadConfig()
		if err == nil {
			t.Fatal("expected error when USERS_DSN is missing, got nil")
		}
	})

	t.Run("uses USERS_DSN when set", func(t *testing.T) {
		withRequiredEnvs(t)
		dsn := "postgres://user:pass@localhost:5432/users"
		t.Setenv("USERS_DSN", dsn)
		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error: %v", err)
		}
		if cfg.DSN != dsn {
			t.Errorf("DSN = %q, want %q", cfg.DSN, dsn)
		}
	})
}

func TestLoadConfig_JWTSecret(t *testing.T) {
	t.Run("errors when USERS_JWT_SECRET is not set", func(t *testing.T) {
		withRequiredEnvs(t)
		os.Unsetenv("USERS_JWT_SECRET")
		_, err := LoadConfig()
		if err == nil {
			t.Fatal("expected error when USERS_JWT_SECRET is missing, got nil")
		}
	})
}

func TestLoadConfig_JWTInternalSecret(t *testing.T) {
	t.Run("errors when USERS_JWT_INTERNAL_SECRET is not set", func(t *testing.T) {
		withRequiredEnvs(t)
		os.Unsetenv("USERS_JWT_INTERNAL_SECRET")
		_, err := LoadConfig()
		if err == nil {
			t.Fatal("expected error when USERS_JWT_INTERNAL_SECRET is missing, got nil")
		}
	})
}

func TestLoadConfig_WalletURL(t *testing.T) {
	t.Run("errors when USERS_WALLET_URL is not set", func(t *testing.T) {
		withRequiredEnvs(t)
		os.Unsetenv("USERS_WALLET_URL")
		_, err := LoadConfig()
		if err == nil {
			t.Fatal("expected error when USERS_WALLET_URL is missing, got nil")
		}
	})
}
