package auth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/youmark/pkcs8"
)

func (a *Authenticator) performClientCredentialsFlow(ctx context.Context, verbose bool) (*TokenResponse, error) {
	// Discover OIDC endpoints
	discovery, err := a.discoverOIDC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OIDC endpoints: %w", err)
	}

	data := url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {strings.Join(a.config.Scopes, " ")},
	}

	if a.config.PrivateKey != "" {
		pemBytes := []byte(a.config.PrivateKey)
		if _, err := os.Stat(a.config.PrivateKey); err == nil {
			bytes, err := os.ReadFile(a.config.PrivateKey)
			if err == nil {
				pemBytes = bytes
			}
		}

		var privateKey *rsa.PrivateKey
		var err error
		if a.config.PrivateKeyPassword != "" {
			block, _ := pem.Decode(pemBytes)
			if block == nil {
				err = fmt.Errorf("failed to parse PEM block containing the key")
			} else {
				var key interface{}
				key, err = pkcs8.ParsePKCS8PrivateKey(block.Bytes, []byte(a.config.PrivateKeyPassword))
				if err == nil {
					var ok bool
					privateKey, ok = key.(*rsa.PrivateKey)
					if !ok {
						err = fmt.Errorf("key is not an RSA private key")
					}
				}
			}
		} else {
			privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(pemBytes)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		claims := jwt.RegisteredClaims{
			Issuer:    a.config.ClientID,
			Subject:   a.config.ClientID,
			Audience:  jwt.ClaimStrings{discovery.TokenEndpoint},
			ID:        uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signedJWT, err := token.SignedString(privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign JWT: %w", err)
		}

		data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		data.Set("client_assertion", signedJWT)
	} else {
		data.Set("client_id", a.config.ClientID)
		data.Set("client_secret", a.config.ClientSecret)
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
