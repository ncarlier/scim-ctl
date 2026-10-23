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
)

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
func (a *Authenticator) ClearCache(verbose bool) error {
	return a.clearCache(verbose)
}

// GetCacheInfo returns information about the cached token (public method)
func (a *Authenticator) GetCacheInfo() (*CachedToken, error) {
	return a.loadCachedToken(false)
}

// hashConfig creates a hash of the configuration for cache file naming
