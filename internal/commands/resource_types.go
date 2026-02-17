package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/idf-educ/idm/scim-ctl/pkg/config"
	"github.com/idf-educ/idm/scim-ctl/pkg/scim"
	"github.com/spf13/cobra"
)

// resourceTypesCmd represents the resource-types command
var resourceTypesCmd = &cobra.Command{
	Use:   "resource-types",
	Short: "List the resource types supported by the server",
	Long: `The resource-types command retrieves and displays all SCIM resource types
supported by the server. Each resource type defines the endpoint, schema,
and schema extensions for a type of resource (e.g. User, Group).`,
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

		resourceTypes, err := client.GetResourceTypes(ctx)
		if err != nil {
			return fmt.Errorf("failed to get resource types: %w", err)
		}

		// Pretty print the resource types
		jsonData, err := json.MarshalIndent(resourceTypes, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal resource types: %w", err)
		}

		fmt.Println(string(jsonData))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resourceTypesCmd)
}
