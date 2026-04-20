package config

import (
	"os"
	"testing"
)

func withRequiredEnvs(t *testing.T) {
	t.Helper()
	t.Setenv("WALLET_DSN", "postgres://user:pass@localhost:5432/wallet")
	t.Setenv("WALLET_PORT", "3001")
	t.Setenv("WALLET_JWT_SECRET", "ILIACHALLENGE")
	t.Setenv("WALLET_JWT_INTERNAL_SECRET", "ILIACHALLENGE_INTERNAL")
}

func TestLoadConfig_Port(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantPort string
		wantErr  bool
	}{
		{
			name:     "uses WALLET_PORT when set",
			envValue: "8080",
			wantPort: "8080",
		},
		{
			name:    "errors when WALLET_PORT is not set",
			wantErr: true,
		},
		{
			name:     "accepts non-standard port",
			envValue: "9999",
			wantPort: "9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withRequiredEnvs(t)

			if tt.envValue != "" {
				t.Setenv("WALLET_PORT", tt.envValue)
			} else {
				os.Unsetenv("WALLET_PORT")
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
				t.Errorf("LoadConfig().Port = %q, want %q", cfg.Port, tt.wantPort)
			}
		})
	}
}

func TestLoadConfig_Release(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		wantRelease bool
	}{
		{
			name:        "defaults to false when WALLET_RELEASE is unset",
			wantRelease: false,
		},
		{
			name:        "sets release to true when WALLET_RELEASE=true",
			envValue:    "true",
			wantRelease: true,
		},
		{
			name:        "sets release to false when WALLET_RELEASE=false",
			envValue:    "false",
			wantRelease: false,
		},
		{
			name:        "sets release to false for invalid value",
			envValue:    "yes",
			wantRelease: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withRequiredEnvs(t)

			if tt.envValue != "" {
				t.Setenv("WALLET_RELEASE", tt.envValue)
			} else {
				os.Unsetenv("WALLET_RELEASE")
			}

			cfg, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() returned error: %v", err)
			}
			if cfg.Release != tt.wantRelease {
				t.Errorf("LoadConfig().Release = %v, want %v", cfg.Release, tt.wantRelease)
			}
		})
	}
}

func TestLoadConfig_DSN(t *testing.T) {
	t.Run("errors when WALLET_DSN is not set", func(t *testing.T) {
		withRequiredEnvs(t)
		os.Unsetenv("WALLET_DSN")
		_, err := LoadConfig()
		if err == nil {
			t.Fatal("expected error when WALLET_DSN is missing, got nil")
		}
	})

	t.Run("uses WALLET_DSN when set", func(t *testing.T) {
		withRequiredEnvs(t)
		dsn := "postgres://user:pass@localhost:5432/wallet"
		t.Setenv("WALLET_DSN", dsn)

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() returned error: %v", err)
		}
		if cfg.DSN != dsn {
			t.Errorf("LoadConfig().DSN = %q, want %q", cfg.DSN, dsn)
		}
	})
}

func TestLoadConfig_JWTSecret(t *testing.T) {
	t.Run("errors when WALLET_JWT_SECRET is not set", func(t *testing.T) {
		withRequiredEnvs(t)
		os.Unsetenv("WALLET_JWT_SECRET")
		_, err := LoadConfig()
		if err == nil {
			t.Fatal("expected error when WALLET_JWT_SECRET is missing, got nil")
		}
	})

	t.Run("uses WALLET_JWT_SECRET when set", func(t *testing.T) {
		withRequiredEnvs(t)
		t.Setenv("WALLET_JWT_SECRET", "mysecret")

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() returned error: %v", err)
		}
		if cfg.JWTSecret != "mysecret" {
			t.Errorf("LoadConfig().JWTSecret = %q, want %q", cfg.JWTSecret, "mysecret")
		}
	})
}

func TestLoadConfig_JWTInternalSecret(t *testing.T) {
	t.Run("errors when WALLET_JWT_INTERNAL_SECRET is not set", func(t *testing.T) {
		withRequiredEnvs(t)
		os.Unsetenv("WALLET_JWT_INTERNAL_SECRET")
		_, err := LoadConfig()
		if err == nil {
			t.Fatal("expected error when WALLET_JWT_INTERNAL_SECRET is missing, got nil")
		}
	})

	t.Run("uses WALLET_JWT_INTERNAL_SECRET when set", func(t *testing.T) {
		withRequiredEnvs(t)
		t.Setenv("WALLET_JWT_INTERNAL_SECRET", "internalsecret")

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() returned error: %v", err)
		}
		if cfg.JWTInternalSecret != "internalsecret" {
			t.Errorf("LoadConfig().JWTInternalSecret = %q, want %q", cfg.JWTInternalSecret, "internalsecret")
		}
	})
}
