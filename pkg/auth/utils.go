package auth

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
