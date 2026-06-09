package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ncarlier/scim-ctl/pkg/config"
	"github.com/ncarlier/scim-ctl/pkg/scim"
	"github.com/spf13/cobra"
)

var (
	updateResourceType string
	updateID           string
	updateData         string
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing resource with a partial update (PATCH)",
	Long: `Update an existing SCIM resource using a PATCH operation. 
The resource data can be provided via the --data flag or through STDIN.
The data must be an array of patch operations.

Examples:
  scim-ctl update --resource user --id 1234 --data '[{"op":"replace","path":"userName","value":"johndoe"}]'
  cat patch.json | scim-ctl update -r user --id 1234`,
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

		// Get data from flag or STDIN
		var dataStr string
		if updateData != "" {
			dataStr = updateData
		} else {
			// Read from STDIN
			stdinData, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from STDIN: %w", err)
			}
			dataStr = string(stdinData)
		}

		if dataStr == "" {
			return fmt.Errorf("operations payload is required (use --data or pipe via STDIN)")
		}

		// Parse JSON data (Array of PatchOperations)
		var operations []scim.PatchOperation
		if err := json.Unmarshal([]byte(dataStr), &operations); err != nil {
			return fmt.Errorf("invalid JSON operations payload: %w", err)
		}

		// Update the resource
		updatedResource, err := client.UpdateResource(ctx, updateResourceType, updateID, operations)
		if err != nil {
			return fmt.Errorf("failed to update resource: %w", err)
		}

		// Pretty print the result
		jsonData, err := json.MarshalIndent(updatedResource, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		fmt.Println(string(jsonData))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(&updateResourceType, "resource", "r", "", "SCIM resource type (required)")
	updateCmd.Flags().StringVar(&updateID, "id", "", "SCIM resource identifier (required)")
	updateCmd.Flags().StringVarP(&updateData, "data", "d", "", "SCIM operations payload (JSON array of operations)")
	updateCmd.MarkFlagRequired("resource")
	updateCmd.MarkFlagRequired("id")
}
