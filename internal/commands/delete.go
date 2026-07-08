package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ncarlier/scim-ctl/pkg/config"
	"github.com/ncarlier/scim-ctl/pkg/scim"
	"github.com/spf13/cobra"
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
  scim-ctl delete --resource user --id 1234
  scim-ctl delete -r group --id abcd-efgh-ijkl`,
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

		result := map[string]string{
			"message":  fmt.Sprintf("Resource %s/%s deleted successfully", deleteResourceType, deleteID),
			"resource": deleteResourceType,
			"id":       deleteID,
		}

		if resultJSON, err := json.Marshal(result); err == nil {
			fmt.Println(string(resultJSON))
		} else {
			fmt.Fprintf(os.Stderr, "Failed to marshal output: %v\n", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&deleteResourceType, "resource", "r", "", "SCIM resource type (required)")
	deleteCmd.Flags().StringVar(&deleteID, "id", "", "SCIM resource identifier (required)")
	deleteCmd.MarkFlagRequired("resource")
	deleteCmd.MarkFlagRequired("id")
}
