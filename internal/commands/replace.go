package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/idf-educ/idm/scim-ctl/pkg/config"
	"github.com/idf-educ/idm/scim-ctl/pkg/scim"
	"github.com/spf13/cobra"
)

var (
	replaceResourceType string
	replaceID           string
	replaceData         string
)

// replaceCmd represents the replace command
var replaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "Replace an existing resource",
	Long: `Replace an existing SCIM resource. The resource data can be provided
via the --data flag or through STDIN.

Examples:
  scim-ctl replace --resource user --id 1234 --data '{"userName": "johndoe"}'
  cat user.json | scim-ctl replace -r user --id 1234`,
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
		if replaceData != "" {
			dataStr = replaceData
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

		// Replace the resource
		replacedResource, err := client.ReplaceResource(ctx, replaceResourceType, replaceID, resourceData)
		if err != nil {
			return fmt.Errorf("failed to replace resource: %w", err)
		}

		// Pretty print the result
		jsonData, err := json.MarshalIndent(replacedResource, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		fmt.Println(string(jsonData))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(replaceCmd)

	replaceCmd.Flags().StringVarP(&replaceResourceType, "resource", "r", "", "SCIM resource type (required)")
	replaceCmd.Flags().StringVar(&replaceID, "id", "", "SCIM resource identifier (required)")
	replaceCmd.Flags().StringVarP(&replaceData, "data", "d", "", "SCIM resource payload (JSON)")
	replaceCmd.MarkFlagRequired("resource")
	replaceCmd.MarkFlagRequired("id")
}
