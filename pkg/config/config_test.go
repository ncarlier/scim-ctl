package config

import (
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "valid config",
			config: Config{
				Target: "https://example.com/scim/v2",
				OIDC: OIDC{
					Issuer:   "https://auth.example.com",
					ClientID: "test-client",
				},
			},
			wantError: false,
		},
		{
			name: "missing target",
			config: Config{
				OIDC: OIDC{
					Issuer:   "https://auth.example.com",
					ClientID: "test-client",
				},
			},
			wantError: true,
		},
		{
			name: "missing OIDC issuer",
			config: Config{
				Target: "https://example.com/scim/v2",
				OIDC: OIDC{
					ClientID: "test-client",
				},
			},
			wantError: true,
		},
		{
			name: "missing OIDC client ID",
			config: Config{
				Target: "https://example.com/scim/v2",
				OIDC: OIDC{
					Issuer: "https://auth.example.com",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Config.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}