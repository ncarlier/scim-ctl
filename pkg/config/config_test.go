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
		{
			name: "missing OIDC client secret for client_credentials",
			config: Config{
				Target: "https://example.com/scim/v2",
				OIDC: OIDC{
					Issuer:    "https://auth.example.com",
					ClientID:  "test-client",
					GrantType: "client_credentials",
				},
			},
			wantError: true,
		},
		{
			name: "valid client_credentials config",
			config: Config{
				Target: "https://example.com/scim/v2",
				OIDC: OIDC{
					Issuer:       "https://auth.example.com",
					ClientID:     "test-client",
					ClientSecret: "test-secret",
					GrantType:    "client_credentials",
				},
			},
			wantError: false,
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
