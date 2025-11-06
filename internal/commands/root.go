package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	target     string
	verbose    bool
	oidcIssuer string
	clientID   string
	clientSecret string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "scim-ctl",
	Short: "A CLI tool for SCIM (System for Cross-domain Identity Management) operations",
	Long: `scim-ctl is a CLI tool for interacting with a SCIM server. It supports CRUD operations
and uses OAuth 2.0 Device Authorization Grant for authentication.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.scim-ctl.yaml)")
	rootCmd.PersistentFlags().StringVar(&target, "target", "", "SCIM target URL (env: SCIM_CTL_TARGET)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&oidcIssuer, "oidc-issuer", "", "OpenID Connect Issuer (env: SCIM_CTL_OIDC_ISSUER)")
	rootCmd.PersistentFlags().StringVar(&clientID, "oidc-client-id", "", "OIDC Client ID (env: SCIM_CTL_OIDC_CLIENT_ID)")
	rootCmd.PersistentFlags().StringVar(&clientSecret, "oidc-client-secret", "", "OIDC Client Secret (env: SCIM_CTL_OIDC_CLIENT_SECRET)")

	// Bind flags to viper
	viper.BindPFlag("target", rootCmd.PersistentFlags().Lookup("target"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("oidc.issuer", rootCmd.PersistentFlags().Lookup("oidc-issuer"))
	viper.BindPFlag("oidc.client-id", rootCmd.PersistentFlags().Lookup("oidc-client-id"))
	viper.BindPFlag("oidc.client-secret", rootCmd.PersistentFlags().Lookup("oidc-client-secret"))

	// Bind environment variables
	viper.BindEnv("target", "SCIM_CTL_TARGET")
	viper.BindEnv("oidc.issuer", "SCIM_CTL_OIDC_ISSUER")
	viper.BindEnv("oidc.client-id", "SCIM_CTL_OIDC_CLIENT_ID")
	viper.BindEnv("oidc.client-secret", "SCIM_CTL_OIDC_CLIENT_SECRET")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".scim-ctl")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}