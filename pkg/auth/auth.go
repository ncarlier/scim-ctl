package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// DeviceFlowConfig represents the device flow configuration
type DeviceFlowConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	Scopes       []string
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

// Authenticator handles OAuth 2.0 Device Authorization Grant flow
type Authenticator struct {
	config *DeviceFlowConfig
	client *http.Client
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator(config *DeviceFlowConfig) *Authenticator {
	return &Authenticator{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetAccessToken performs the device flow and returns an access token
func (a *Authenticator) GetAccessToken(ctx context.Context, verbose bool) (string, error) {
	// Discover OIDC endpoints
	discovery, err := a.discoverOIDC(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to discover OIDC endpoints: %w", err)
	}

	// Request device code
	deviceCode, err := a.requestDeviceCode(ctx, discovery.DeviceAuthorizationEndpoint, verbose)
	if err != nil {
		return "", fmt.Errorf("failed to request device code: %w", err)
	}

	// Poll for token
	token, err := a.pollForToken(ctx, discovery.TokenEndpoint, deviceCode, verbose)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	return token.AccessToken, nil
}

// discoverOIDC discovers OIDC endpoints
func (a *Authenticator) discoverOIDC(ctx context.Context) (*OIDCDiscovery, error) {
	discoveryURL := strings.TrimSuffix(a.config.Issuer, "/") + "/.well-known/openid-configuration"
	
	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery request failed with status %d", resp.StatusCode)
	}

	var discovery OIDCDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, err
	}

	return &discovery, nil
}

// requestDeviceCode requests a device code
func (a *Authenticator) requestDeviceCode(ctx context.Context, endpoint string, verbose bool) (*DeviceCodeResponse, error) {
	data := url.Values{
		"client_id": {a.config.ClientID},
		"scope":     {strings.Join(a.config.Scopes, " ")},
	}

	if a.config.ClientSecret != "" {
		data.Set("client_secret", a.config.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if verbose {
		fmt.Fprintf(os.Stderr, "Requesting device code from %s\n", endpoint)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed with status %d", resp.StatusCode)
	}

	var deviceCode DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceCode); err != nil {
		return nil, err
	}

	// Display user instructions
	fmt.Printf("Please visit %s and enter the code: %s\n", deviceCode.VerificationURI, deviceCode.UserCode)
	if deviceCode.VerificationURIComplete != "" {
		fmt.Printf("Or visit: %s\n", deviceCode.VerificationURIComplete)
	}

	return &deviceCode, nil
}

// pollForToken polls for an access token
func (a *Authenticator) pollForToken(ctx context.Context, endpoint string, deviceCode *DeviceCodeResponse, verbose bool) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"device_code": {deviceCode.DeviceCode},
		"client_id":   {a.config.ClientID},
	}

	if a.config.ClientSecret != "" {
		data.Set("client_secret", a.config.ClientSecret)
	}

	interval := time.Duration(deviceCode.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	timeout := time.Duration(deviceCode.ExpiresIn) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if verbose {
		fmt.Fprintf(os.Stderr, "Polling for token every %v\n", interval)
	}

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("device code expired")
		case <-ticker.C:
			req, err := http.NewRequestWithContext(timeoutCtx, "POST", endpoint, strings.NewReader(data.Encode()))
			if err != nil {
				return nil, err
			}

			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := a.client.Do(req)
			if err != nil {
				return nil, err
			}

			var tokenResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&tokenResp)
			resp.Body.Close()

			if err != nil {
				return nil, err
			}

			if resp.StatusCode == http.StatusOK {
				token := &TokenResponse{
					AccessToken:  tokenResp["access_token"].(string),
					TokenType:    tokenResp["token_type"].(string),
					RefreshToken: getString(tokenResp, "refresh_token"),
					Scope:        getString(tokenResp, "scope"),
				}
				if expiresIn, ok := tokenResp["expires_in"].(float64); ok {
					token.ExpiresIn = int(expiresIn)
				}
				return token, nil
			}

			// Handle specific error cases
			if errorCode := getString(tokenResp, "error"); errorCode != "" {
				switch errorCode {
				case "authorization_pending":
					// Continue polling
					if verbose {
						fmt.Fprintf(os.Stderr, "Authorization pending...\n")
					}
					continue
				case "slow_down":
					// Increase polling interval
					interval += time.Second
					ticker.Reset(interval)
					continue
				case "expired_token":
					return nil, fmt.Errorf("device code expired")
				case "access_denied":
					return nil, fmt.Errorf("access denied by user")
				default:
					return nil, fmt.Errorf("token request failed: %s", errorCode)
				}
			}
		}
	}
}

// ValidateJWT validates a JWT token (basic validation)
func ValidateJWT(tokenString string) (*jwt.Token, error) {
	// Parse token without verification for now
	// In production, you should verify the signature with the proper key
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Basic validation of claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check if token is expired
		if exp, ok := claims["exp"]; ok {
			if expTime, ok := exp.(float64); ok {
				if time.Unix(int64(expTime), 0).Before(time.Now()) {
					return nil, fmt.Errorf("token is expired")
				}
			}
		}
	}

	return token, nil
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}