package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ncarlier/scim-ctl/pkg/config"
	"github.com/ncarlier/scim-ctl/pkg/scim"
	"github.com/spf13/cobra"
)

var (
	exportResourceType string
	exportFilter       string
	exportQuery        string
	exportItemsPerPage int
	exportSortBy       string
	exportSortOrder          string
	exportAttributes         []string
	exportExcludedAttributes []string
	exportUseCursor          bool
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
  scim-ctl export -r user -f 'active eq true' --attributes userName,emails
  scim-ctl export -r user --use-cursor`,
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

		// Detect if stdout is being redirected or piped
		fileInfo, _ := os.Stdout.Stat()
		isRedirected := (fileInfo.Mode() & os.ModeCharDevice) == 0

		startIndex := 1
		exportedCount := 0
		lastReportedPercent := -1
		totalResults := 0
		isFirstPage := true

		var currentCursor *string
		if exportUseCursor {
			emptyCursor := ""
			currentCursor = &emptyCursor
		}

		var count *int
		if cmd.Flags().Changed("items-per-page") {
			count = &exportItemsPerPage
		}

		startTime := time.Now()

		for {
			// Search for resources
			results, err := client.SearchResources(ctx, exportResourceType, exportFilter, exportQuery, startIndex, count, exportSortBy, exportSortOrder, exportAttributes, exportExcludedAttributes, currentCursor)
			if err != nil {
				if exportUseCursor {
					return fmt.Errorf("failed to search resources with cursor: %w", err)
				}
				return fmt.Errorf("failed to search resources at start index %d: %w", startIndex, err)
			}

			if isFirstPage {
				if results.TotalResults > 0 {
					totalResults = results.TotalResults
					if isRedirected {
						fmt.Fprintf(os.Stderr, "Total resources to export: %d\n", totalResults)
					}
				}
				isFirstPage = false
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

				exportedCount++
				
				if isRedirected {
					if totalResults > 0 {
						currentPercent := (exportedCount * 100) / totalResults
						if currentPercent > lastReportedPercent {
							elapsed := time.Since(startTime)
							eta := time.Duration(float64(elapsed) / float64(exportedCount) * float64(totalResults-exportedCount)).Round(time.Second)
							fmt.Fprintf(os.Stderr, "Exporting... %d%% (ETA: %v)\033[K\r", currentPercent, eta)
							lastReportedPercent = currentPercent
						}
					} else {
						elapsed := time.Since(startTime).Round(time.Second)
						fmt.Fprintf(os.Stderr, "Exporting... %d resources (Elapsed: %v)\033[K\r", exportedCount, elapsed)
					}
				}
			}

			if exportUseCursor {
				if results.NextCursor == nil || *results.NextCursor == "" {
					break
				}
				currentCursor = results.NextCursor
			} else {
				// Increment startIndex for the next page
				startIndex += len(results.Resources)
			}
		}

		if isRedirected {
			fmt.Fprintln(os.Stderr)
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
	exportCmd.Flags().StringSliceVarP(&exportExcludedAttributes, "excluded-attributes", "e", []string{}, "Comma-separated list of attributes to exclude")
	exportCmd.Flags().BoolVar(&exportUseCursor, "use-cursor", false, "Use cursor-based pagination (RFC 9865)")
	exportCmd.MarkFlagRequired("resource")
}
