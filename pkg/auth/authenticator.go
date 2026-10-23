package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"encoding/json"
	"time"

	"github.com/ncarlier/scim-ctl/pkg/common"
)

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

	// Set default timeout if not provided
	timeout := config.Timeout
	if timeout == 0 {
		timeout = common.DefaultHTTPClientTimeout
	}

	return &Authenticator{
		config: config,
		client: &http.Client{
			Timeout: timeout,
			Transport: &common.RetryTransport{
				Transport:  http.DefaultTransport,
				MaxRetries: config.MaxRetries,
				Verbose:    config.Verbose,
			},
		},
		cacheFile: cacheFile,
	}
}

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

