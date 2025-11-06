package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/idf-educ/idm/scim-ctl/pkg/config"
	"github.com/idf-educ/idm/scim-ctl/pkg/scim"
)

var (
	updateResourceType string
	updateID           string
	updateData         string
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing resource",
	Long: `Update an existing SCIM resource. The resource data can be provided
via the --data flag or through STDIN.

Examples:
  scim-ctl update --resource-type user --id 1234 --data '{"userName": "johndoe"}'
  cat user.json | scim-ctl update -t user --id 1234`,
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
			return fmt.Errorf("resource data is required (use --data or pipe via STDIN)")
		}

		// Parse JSON data
		var resourceData scim.Resource
		if err := json.Unmarshal([]byte(dataStr), &resourceData); err != nil {
			return fmt.Errorf("invalid JSON data: %w", err)
		}

		// Update the resource
		updatedResource, err := client.UpdateResource(ctx, updateResourceType, updateID, resourceData)
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
	
	updateCmd.Flags().StringVarP(&updateResourceType, "resource-type", "t", "", "SCIM resource type (required)")
	updateCmd.Flags().StringVar(&updateID, "id", "", "SCIM resource identifier (required)")
	updateCmd.Flags().StringVarP(&updateData, "data", "d", "", "SCIM resource payload (JSON)")
	updateCmd.MarkFlagRequired("resource-type")
	updateCmd.MarkFlagRequired("id")
}