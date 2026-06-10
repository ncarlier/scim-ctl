package auth

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Default valifdity time window for tokens
const TOKEN_VALIDITY_WINDOW = 1 * time.Minute

// AuthConfig represents the authentication configuration
type AuthConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	GrantType    string
	Scopes       []string
	CacheDir     string
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

// Authenticator handles OAuth 2.0 Device Authorization Grant flow
type Authenticator struct {
	config    *AuthConfig
	client    *http.Client
	cacheFile string
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator(config *AuthConfig) *Authenticator {
	// Create cache file path based on configuration
	var cacheDir string
	if config.CacheDir != "" {
		cacheDir = config.CacheDir
	} else {
		homeDir, _ := os.UserHomeDir()
		cacheDir = fmt.Sprintf("%s/.cache/scim-ctl", homeDir)
	}
	os.MkdirAll(cacheDir, 0700) // Create cache directory with restricted permissions

	// Create a unique cache file name based on issuer and client_id
	cacheFile := fmt.Sprintf("%s/token_%x.json", cacheDir, hashConfig(config))

	return &Authenticator{
		config:    config,
		client:    &http.Client{Timeout: 30 * time.Second},
		cacheFile: cacheFile,
	}
}

// loadCachedToken loads a token from cache if it exists and is valid
func (a *Authenticator) loadCachedToken(verbose bool) (*CachedToken, error) {
	data, err := os.ReadFile(a.cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No cached token
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cached CachedToken
	if err := json.Unmarshal(data, &cached); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Invalid cache file, ignoring: %v\n", err)
		}
		return nil, nil
	}

	// Check if token is for the same configuration
	if cached.Issuer != a.config.Issuer || cached.ClientID != a.config.ClientID || cached.GrantType != a.config.GrantType {
		if verbose {
			fmt.Fprintf(os.Stderr, "Cache token is for different configuration, ignoring\n")
		}
		return nil, nil
	}

	return &cached, nil
}

// saveCachedToken saves a token to cache
func (a *Authenticator) saveCachedToken(token *TokenResponse, verbose bool) error {
	// Calculate expiry time
	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	cached := CachedToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    expiresAt,
		Scopes:       token.Scope,
		Issuer:       a.config.Issuer,
		ClientID:     a.config.ClientID,
		GrantType:    a.config.GrantType,
	}

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(a.cacheFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Token cached to %s\n", a.cacheFile)
	}

	return nil
}

// refreshToken attempts to refresh an access token using the refresh token
func (a *Authenticator) refreshToken(ctx context.Context, cachedToken *CachedToken, verbose bool) (*TokenResponse, error) {
	if cachedToken.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Discover OIDC endpoints
	discovery, err := a.discoverOIDC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OIDC endpoints: %w", err)
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {cachedToken.RefreshToken},
		"client_id":     {a.config.ClientID},
	}

	if a.config.ClientSecret != "" {
		data.Set("client_secret", a.config.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", discovery.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if verbose {
		fmt.Fprintf(os.Stderr, "Refreshing token...\n")
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status %d: %s",
			resp.StatusCode, getString(tokenResp, "error_description"))
	}

	token := &TokenResponse{
		AccessToken:  getString(tokenResp, "access_token"),
		TokenType:    getString(tokenResp, "token_type"),
		RefreshToken: getString(tokenResp, "refresh_token"),
		Scope:        getString(tokenResp, "scope"),
	}

	// Use existing refresh token if new one not provided
	if token.RefreshToken == "" {
		token.RefreshToken = cachedToken.RefreshToken
	}

	if expiresIn, ok := tokenResp["expires_in"].(float64); ok {
		token.ExpiresIn = int(expiresIn)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Token refreshed successfully\n")
	}

	return token, nil
}

// clearCache removes the cached token file
func (a *Authenticator) clearCache(verbose bool) error {
	if err := os.Remove(a.cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Token cache cleared\n")
	}

	return nil
}

// GetAccessTokenSilent returns the cached or refreshed token without starting a new device flow.
func (a *Authenticator) GetAccessTokenSilent(ctx context.Context, verbose bool) (string, error) {
	// Try to load cached token first
	cachedToken, err := a.loadCachedToken(verbose)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Failed to load cached token: %v\n", err)
		}
		return "", err
	}

	// Check if we have a valid cached token
	if cachedToken != nil {
		// Check if token is still valid (with 1 minute buffer)
		if time.Now().Add(TOKEN_VALIDITY_WINDOW).Before(cachedToken.ExpiresAt) {
			if verbose {
				fmt.Fprintf(os.Stderr, "Using cached access token (expires: %v)\n", cachedToken.ExpiresAt.Format(time.RFC3339))
			}
			return cachedToken.AccessToken, nil
		}

		// Token is expired or about to expire, try to refresh
		if cachedToken.RefreshToken != "" {
			if verbose {
				fmt.Fprintf(os.Stderr, "Access token expired, attempting refresh...\n")
			}

			refreshed, err := a.refreshToken(ctx, cachedToken, verbose)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "Token refresh failed: %v\n", err)
				}
				// Clear invalid cache
				a.clearCache(verbose)
				return "", fmt.Errorf("token refresh failed: %w", err)
			}

			// Save refreshed token and return it
			if err := a.saveCachedToken(refreshed, verbose); err != nil && verbose {
				fmt.Fprintf(os.Stderr, "Failed to cache refreshed token: %v\n", err)
			}
			return refreshed.AccessToken, nil
		}

		// If client credentials flow, request a new token silently
		if a.config.GrantType == "client_credentials" {
			if verbose {
				fmt.Fprintf(os.Stderr, "Access token expired, requesting new client credentials token...\n")
			}
			newToken, err := a.performClientCredentialsFlow(ctx, verbose)
			if err != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "Client credentials token request failed: %v\n", err)
				}
				a.clearCache(verbose)
				return "", fmt.Errorf("client credentials token request failed: %w", err)
			}

			if err := a.saveCachedToken(newToken, verbose); err != nil && verbose {
				fmt.Fprintf(os.Stderr, "Failed to cache token: %v\n", err)
			}
			return newToken.AccessToken, nil
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "Access token expired and no refresh token available\n")
		}
		return "", fmt.Errorf("access token expired and no refresh token available")
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "No cached token found\n")
	}
	return "", fmt.Errorf("no cached token found")
}

// GetAccessToken performs the device flow and returns an access token
func (a *Authenticator) GetAccessToken(ctx context.Context, verbose bool) (string, error) {
	// Try to get token silently first
	token, err := a.GetAccessTokenSilent(ctx, verbose)
	if err == nil {
		return token, nil
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Silent authentication failed: %v, performing full authentication flow\n", err)
	}

	var tokenResp *TokenResponse
	if a.config.GrantType == "client_credentials" {
		tokenResp, err = a.performClientCredentialsFlow(ctx, verbose)
	} else {
		// Default to device flow
		tokenResp, err = a.performDeviceFlow(ctx, verbose)
	}

	if err != nil {
		return "", err
	}

	// Cache the new token
	if err := a.saveCachedToken(tokenResp, verbose); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Failed to cache token: %v\n", err)
	}

	return tokenResp.AccessToken, nil
}

// performDeviceFlow executes the OAuth 2.0 device authorization grant flow
func (a *Authenticator) performDeviceFlow(ctx context.Context, verbose bool) (*TokenResponse, error) {
	// Discover OIDC endpoints
	discovery, err := a.discoverOIDC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OIDC endpoints: %w", err)
	}

	// Request device code
	deviceCode, err := a.requestDeviceCode(ctx, discovery.DeviceAuthorizationEndpoint, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}

	// Poll for token
	token, err := a.pollForToken(ctx, discovery.TokenEndpoint, deviceCode, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return token, nil
}

// performClientCredentialsFlow executes the OAuth 2.0 client credentials grant flow
func (a *Authenticator) performClientCredentialsFlow(ctx context.Context, verbose bool) (*TokenResponse, error) {
	// Discover OIDC endpoints
	discovery, err := a.discoverOIDC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OIDC endpoints: %w", err)
	}

	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {a.config.ClientID},
		"client_secret": {a.config.ClientSecret},
		"scope":         {strings.Join(a.config.Scopes, " ")},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", discovery.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if verbose {
		fmt.Fprintf(os.Stderr, "Requesting client credentials token from %s\n", discovery.TokenEndpoint)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client credentials request failed with status %d: %s",
			resp.StatusCode, getString(tokenResp, "error_description"))
	}

	token := &TokenResponse{
		AccessToken: getString(tokenResp, "access_token"),
		TokenType:   getString(tokenResp, "token_type"),
		Scope:       getString(tokenResp, "scope"),
	}

	if expiresIn, ok := tokenResp["expires_in"].(float64); ok {
		token.ExpiresIn = int(expiresIn)
	}

	return token, nil
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
					if verbose {
						fmt.Fprintf(os.Stderr, "Token received (expires in %d seconds)\n", token.ExpiresIn)
					}
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

// ClearCache removes the cached token file (public method)
func (a *Authenticator) ClearCache(verbose bool) error {
	return a.clearCache(verbose)
}

// GetCacheInfo returns information about the cached token (public method)
func (a *Authenticator) GetCacheInfo() (*CachedToken, error) {
	return a.loadCachedToken(false)
}

// hashConfig creates a hash of the configuration for cache file naming
func hashConfig(config *AuthConfig) [32]byte {
	data := fmt.Sprintf("%s:%s:%s", config.Issuer, config.ClientID, config.GrantType)
	return sha256.Sum256([]byte(data))
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
