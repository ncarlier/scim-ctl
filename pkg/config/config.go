package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Target       string            `mapstructure:"target"`
	Verbose      bool              `mapstructure:"verbose"`
	OIDC         OIDC              `mapstructure:"oidc"`
	ExtraHeaders map[string]string `mapstructure:"extra-headers"`
}

// OIDC represents the OAuth 2.0 configuration
type OIDC struct {
	Issuer       string `mapstructure:"issuer"`
	ClientID     string `mapstructure:"client-id"`
	ClientSecret string `mapstructure:"client-secret"`
}

// Get returns the current configuration
func Get() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Target == "" {
		return errors.New("SCIM target URL is required")
	}

	if c.OIDC.Issuer == "" {
		return errors.New("OIDC issuer is required")
	}

	if c.OIDC.ClientID == "" {
		return errors.New("OIDC client ID is required")
	}

	return nil
}