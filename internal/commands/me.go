package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ncarlier/scim-ctl/pkg/config"
	"github.com/ncarlier/scim-ctl/pkg/scim"
	"github.com/spf13/cobra"
)

var meAttributes []string

// meCmd represents the me command
var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Retrieve the authenticated user's resource",
	Long: `Retrieve the SCIM resource for the currently authenticated user using the /Me endpoint.

Examples:
  scim-ctl me
  scim-ctl me --attributes userName,emails
  scim-ctl me -a userName -a emails`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Get()
		if err != nil {
			return fmt.Errorf("failed to get configuration: %w", err)
		}

		client, err := scim.NewClient(cfg)
		if err != nil {
			return fmt.Errorf("failed to create SCIM client: %w", err)
		}

		ctx := context.Background()
		if err := client.Authenticate(ctx, cfg); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Get the current user
		resource, err := client.GetResource(ctx, "Me", "", meAttributes)
		if err != nil {
			return fmt.Errorf("failed to get /Me resource: %w", err)
		}

		// Pretty print the result
		jsonData, err := json.MarshalIndent(resource, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		fmt.Println(string(jsonData))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(meCmd)
	meCmd.Flags().StringSliceVarP(&meAttributes, "attributes", "a", []string{}, "Comma-separated list of attributes to return")
}
