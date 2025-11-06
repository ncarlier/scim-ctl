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
	createResourceType string
	createData         string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a SCIM resource",
	Long: `Create a SCIM resource on the server. The resource data can be provided
via the --data flag or through STDIN.

Examples:
  scim-ctl create --resource-type user --data '{"userName": "jdoe", "emails": [{"value": "jdoe@example.com"}]}'
  cat user.json | scim-ctl create -t user`,
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
		if createData != "" {
			dataStr = createData
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

		// Create the resource
		createdResource, err := client.CreateResource(ctx, createResourceType, resourceData)
		if err != nil {
			return fmt.Errorf("failed to create resource: %w", err)
		}

		// Pretty print the result
		jsonData, err := json.MarshalIndent(createdResource, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		fmt.Println(string(jsonData))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	
	createCmd.Flags().StringVarP(&createResourceType, "resource-type", "t", "", "SCIM resource type (required)")
	createCmd.Flags().StringVarP(&createData, "data", "d", "", "SCIM resource payload (JSON)")
	createCmd.MarkFlagRequired("resource-type")
}