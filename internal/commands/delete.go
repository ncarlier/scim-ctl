package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/idf-educ/idm/scim-ctl/pkg/config"
	"github.com/idf-educ/idm/scim-ctl/pkg/scim"
)

var (
	deleteResourceType string
	deleteID           string
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a SCIM resource",
	Long: `Delete a SCIM resource by its unique identifier.

Examples:
  scim-ctl delete --resource-type user --id 1234
  scim-ctl delete -t group --id abcd-efgh-ijkl`,
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

		// Delete the resource
		if err := client.DeleteResource(ctx, deleteResourceType, deleteID); err != nil {
			return fmt.Errorf("failed to delete resource: %w", err)
		}

		fmt.Printf("Resource %s/%s deleted successfully\n", deleteResourceType, deleteID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	
	deleteCmd.Flags().StringVarP(&deleteResourceType, "resource-type", "t", "", "SCIM resource type (required)")
	deleteCmd.Flags().StringVar(&deleteID, "id", "", "SCIM resource identifier (required)")
	deleteCmd.MarkFlagRequired("resource-type")
	deleteCmd.MarkFlagRequired("id")
}