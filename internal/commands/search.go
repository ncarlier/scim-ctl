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
	searchResourceType string
	searchFilter       string
	searchQuery        string
	searchStartIndex   int
	searchItemsPerPage int
	searchSortBy       string
	searchSortOrder    string
	searchAttributes   []string
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search SCIM resources",
	Long: `Search for SCIM resources using SCIM filter expressions.

Examples:
  scim-ctl search --resource user --filter 'userName eq "bob"'
  scim-ctl search -r user -q "john doe"
  scim-ctl search -r group -f 'displayName co "admin"' --start-index 1 --items-per-page 10
  scim-ctl search -r user -f 'active eq true' --sort-by userName --sort-order ascending
  scim-ctl search -r user --sort-by meta.created --sort-order descending
  scim-ctl search -r user -f 'active eq true' --attributes userName,emails
  scim-ctl search -r user --attributes userName --attributes emails`,
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

		// Search for resources
		results, err := client.SearchResources(ctx, searchResourceType, searchFilter, searchQuery, searchStartIndex, searchItemsPerPage, searchSortBy, searchSortOrder, searchAttributes)
		if err != nil {
			return fmt.Errorf("failed to search resources: %w", err)
		}

		// Pretty print the results
		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		fmt.Println(string(jsonData))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&searchResourceType, "resource", "r", "", "SCIM resource type (required)")
	searchCmd.Flags().StringVarP(&searchFilter, "filter", "f", "", "SCIM filter expression")
	searchCmd.Flags().StringVarP(&searchQuery, "query", "q", "", "Full-text search query (custom extension)")
	searchCmd.Flags().IntVarP(&searchStartIndex, "start-index", "s", 0, "Pagination start index")
	searchCmd.Flags().IntVarP(&searchItemsPerPage, "items-per-page", "i", 0, "Pagination size")
	searchCmd.Flags().StringVar(&searchSortBy, "sort-by", "", "Attribute to sort by (e.g., userName, meta.created)")
	searchCmd.Flags().StringVar(&searchSortOrder, "sort-order", "", "Sort order: ascending or descending")
	searchCmd.Flags().StringSliceVarP(&searchAttributes, "attributes", "a", []string{}, "Comma-separated list of attributes to return")
	searchCmd.MarkFlagRequired("resource")
}
