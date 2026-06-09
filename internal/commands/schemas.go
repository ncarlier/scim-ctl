package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ncarlier/scim-ctl/pkg/config"
	"github.com/ncarlier/scim-ctl/pkg/scim"
	"github.com/spf13/cobra"
)

// schemasCmd represents the schemas command
var schemasCmd = &cobra.Command{
	Use:   "schemas",
	Short: "Display the resources and attribute extensions supported by the server",
	Long: `The schemas command retrieves and displays all SCIM schemas supported by the server.
This includes both resource schemas (like User, Group) and extensions.`,
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

		schemas, err := client.GetSchemas(ctx)
		if err != nil {
			return fmt.Errorf("failed to get schemas: %w", err)
		}

		// Pretty print the schemas
		jsonData, err := json.MarshalIndent(schemas, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal schemas: %w", err)
		}

		fmt.Println(string(jsonData))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(schemasCmd)
}
