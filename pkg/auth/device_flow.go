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
