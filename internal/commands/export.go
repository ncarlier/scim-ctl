package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/idf-educ/idm/scim-ctl/pkg/config"
	"github.com/idf-educ/idm/scim-ctl/pkg/scim"
	"github.com/spf13/cobra"
)

var (
	exportResourceType string
	exportFilter       string
	exportQuery        string
	exportItemsPerPage int
	exportSortBy       string
	exportSortOrder    string
	exportAttributes   []string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export all SCIM resources as JSON Lines",
	Long: `Export all SCIM resources matching the search criteria.
The output is formatted as JSON Lines (JSONL), with one JSON object per line.
Pagination is handled automatically until all resources are retrieved.

Examples:
  scim-ctl export --resource user --filter 'userName eq "bob"'
  scim-ctl export -r group -f 'displayName co "admin"' --items-per-page 100
  scim-ctl export -r user --sort-by meta.created --sort-order descending
  scim-ctl export -r user -f 'active eq true' --attributes userName,emails`,
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

		startIndex := 1

		for {
			// Search for resources
			results, err := client.SearchResources(ctx, exportResourceType, exportFilter, exportQuery, startIndex, exportItemsPerPage, exportSortBy, exportSortOrder, exportAttributes)
			if err != nil {
				return fmt.Errorf("failed to search resources at start index %d: %w", startIndex, err)
			}

			if len(results.Resources) == 0 {
				break
			}

			// Output each resource as a single JSON line
			for _, res := range results.Resources {
				jsonData, err := json.Marshal(res)
				if err != nil {
					return fmt.Errorf("failed to marshal resource: %w", err)
				}
				fmt.Println(string(jsonData))
			}

			// Increment startIndex for the next page
			startIndex += len(results.Resources)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportResourceType, "resource", "r", "", "SCIM resource type (required)")
	exportCmd.Flags().StringVarP(&exportFilter, "filter", "f", "", "SCIM filter expression")
	exportCmd.Flags().StringVarP(&exportQuery, "query", "q", "", "Full-text search query (custom extension)")
	exportCmd.Flags().IntVarP(&exportItemsPerPage, "items-per-page", "i", 0, "Pagination size")
	exportCmd.Flags().StringVar(&exportSortBy, "sort-by", "", "Attribute to sort by (e.g., userName, meta.created)")
	exportCmd.Flags().StringVar(&exportSortOrder, "sort-order", "", "Sort order: ascending or descending")
	exportCmd.Flags().StringSliceVarP(&exportAttributes, "attributes", "a", []string{}, "Comma-separated list of attributes to return")
	exportCmd.MarkFlagRequired("resource")
}
