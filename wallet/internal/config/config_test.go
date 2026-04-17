package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Port(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		wantPort string
	}{
		{
			name:     "uses WALLET_PORT when set",
			envValue: "8080",
			wantPort: "8080",
		},
		{
			name:     "defaults to 3001 when WALLET_PORT is empty",
			envValue: "",
			wantPort: "3001",
		},
		{
			name:     "accepts non-standard port",
			envValue: "9999",
			wantPort: "9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv("WALLET_PORT", tt.envValue)
			} else {
				os.Unsetenv("WALLET_PORT")
			}

			cfg := LoadConfig()

			if cfg == nil {
				t.Fatal("LoadConfig() returned nil")
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
			envValue:    "",
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
			if tt.envValue != "" {
				t.Setenv("WALLET_RELEASE", tt.envValue)
			} else {
				os.Unsetenv("WALLET_RELEASE")
			}

			cfg := LoadConfig()

			if cfg == nil {
				t.Fatal("LoadConfig() returned nil")
			}
			if cfg.Release != tt.wantRelease {
				t.Errorf("LoadConfig().Release = %v, want %v", cfg.Release, tt.wantRelease)
			}
		})
	}
}
