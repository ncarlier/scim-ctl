package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/idf-educ/idm/scim-ctl/pkg/config"
	"github.com/idf-educ/idm/scim-ctl/pkg/scim"
)

var (
	getResourceType string
	getID           string
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve a resource by ID",
	Long: `Retrieve a SCIM resource by its unique identifier.

Examples:
  scim-ctl get --resource-type user --id 1234
  scim-ctl get -t group --id abcd-efgh-ijkl`,
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

		// Get the resource
		resource, err := client.GetResource(ctx, getResourceType, getID)
		if err != nil {
			return fmt.Errorf("failed to get resource: %w", err)
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
	rootCmd.AddCommand(getCmd)
	
	getCmd.Flags().StringVarP(&getResourceType, "resource-type", "t", "", "SCIM resource type (required)")
	getCmd.Flags().StringVar(&getID, "id", "", "SCIM resource identifier (required)")
	getCmd.MarkFlagRequired("resource-type")
	getCmd.MarkFlagRequired("id")
}