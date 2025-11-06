package commands

import (
	"fmt"
	"time"

	"github.com/idf-educ/idm/scim-ctl/pkg/auth"
	"github.com/idf-educ/idm/scim-ctl/pkg/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheInfoCmd)
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage authentication token cache",
	Long:  `Manage authentication token cache for the SCIM CLI.`,
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear cached authentication tokens",
	Long:  `Clear all cached authentication tokens. This will force re-authentication on the next command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Get()
		if err != nil {
			return fmt.Errorf("failed to get configuration: %w", err)
		}

		authConfig := &auth.DeviceFlowConfig{
			Issuer:       cfg.OIDC.Issuer,
			ClientID:     cfg.OIDC.ClientID,
			ClientSecret: cfg.OIDC.ClientSecret,
			Scopes:       []string{"openid", "profile"},
		}

		authenticator := auth.NewAuthenticator(authConfig)
		
		if err := authenticator.ClearCache(cfg.Verbose); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}

		fmt.Println("Authentication cache cleared successfully")
		return nil
	},
}

var cacheInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show cache information",
	Long:  `Show information about the current authentication token cache.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Get()
		if err != nil {
			return fmt.Errorf("failed to get configuration: %w", err)
		}

		authConfig := &auth.DeviceFlowConfig{
			Issuer:       cfg.OIDC.Issuer,
			ClientID:     cfg.OIDC.ClientID,
			ClientSecret: cfg.OIDC.ClientSecret,
			Scopes:       []string{"openid", "profile"},
		}

		authenticator := auth.NewAuthenticator(authConfig)
		
		// Try to load cached token to show info
		cachedToken, err := authenticator.GetCacheInfo()
		if err != nil {
			return fmt.Errorf("failed to get cache info: %w", err)
		}

		if cachedToken == nil {
			fmt.Println("No cached authentication token found")
			return nil
		}

		fmt.Printf("Cached Token Information:\n")
		fmt.Printf("  Issuer: %s\n", cachedToken.Issuer)
		fmt.Printf("  Client ID: %s\n", cachedToken.ClientID)
		fmt.Printf("  Token Type: %s\n", cachedToken.TokenType)
		fmt.Printf("  Scopes: %s\n", cachedToken.Scopes)
		fmt.Printf("  Expires At: %s\n", cachedToken.ExpiresAt.Format("2006-01-02 15:04:05 MST"))
		
		if cachedToken.ExpiresAt.Before(time.Now()) {
			fmt.Printf("  Status: Expired\n")
		} else {
			fmt.Printf("  Status: Valid\n")
		}
		
		if cachedToken.RefreshToken != "" {
			fmt.Printf("  Refresh Token: Available\n")
		} else {
			fmt.Printf("  Refresh Token: Not available\n")
		}

		return nil
	},
}