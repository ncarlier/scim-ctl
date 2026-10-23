package auth

import (
	"time"
)

// Default valifdity time window for tokens
const TOKEN_VALIDITY_WINDOW = 1 * time.Minute

// AuthConfig represents the authentication configuration
type AuthConfig struct {
	Issuer             string
	ClientID           string
	ClientSecret       string
	PrivateKey         string
	PrivateKeyPassword string
	GrantType          string
	Scopes             []string
	CacheDir           string
	Timeout            time.Duration
	MaxRetries         int
	Verbose            bool
}

// DeviceCodeResponse represents the device code response
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// OIDCDiscovery represents OIDC discovery document
type OIDCDiscovery struct {
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
}

// CachedToken represents a cached token with metadata
type CachedToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scopes       string    `json:"scopes"`
	Issuer       string    `json:"issuer"`
	ClientID     string    `json:"client_id"`
	GrantType    string    `json:"grant_type"`
}
