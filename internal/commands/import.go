package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/ncarlier/scim-ctl/pkg/config"
	"github.com/ncarlier/scim-ctl/pkg/scim"
	"github.com/spf13/cobra"
)

var (
	importResourceType string
	importFile         string
	importChunkSize    int
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import SCIM resources using the Bulk API",
	Long: `Import SCIM resources on the server using the Bulk API.
The input data should be a stream of JSON Lines, where each line is the payload of a resource to create.
The data can be provided via the --file flag or through STDIN.

Examples:
  scim-ctl import --resource user --file users.jsonl
  cat users.jsonl | scim-ctl import -r user --chunk 50`,
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

		// Setup input stream
		var reader io.Reader
		totalLines := 0

		if importFile != "" {
			// Count total lines for progress tracking
			file, err := os.Open(importFile)
			if err != nil {
				return fmt.Errorf("failed to open input file: %w", err)
			}
			totalLines = countLines(file)
			file.Close()

			// Reopen for actual reading
			file, err = os.Open(importFile)
			if err != nil {
				return fmt.Errorf("failed to open input file: %w", err)
			}
			defer file.Close()
			reader = file
		} else {
			// Check if data is piped via STDIN
			stat, err := os.Stdin.Stat()
			if err != nil {
				return fmt.Errorf("failed to check STDIN: %w", err)
			}
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return fmt.Errorf("input data is required (use --file or pipe via STDIN)")
			}
			reader = os.Stdin
		}

		scanner := bufio.NewScanner(reader)
		// Increase scanner buffer size if dealing with large JSON lines
		const maxCapacity = 10 * 1024 * 1024 // 10MB
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		stderrStat, _ := os.Stderr.Stat()
		isStderrTerminal := (stderrStat.Mode() & os.ModeCharDevice) != 0

		var batch []scim.BulkOperation
		var totalSuccess, totalErrors int
		processedCount := 0
		lastReportedPercent := -1
		lineNum := 0

		resourcePath := "/" + scim.ResourceName(importResourceType)

		if isStderrTerminal && totalLines > 0 {
			fmt.Fprintf(os.Stderr, "Total resources to import: %d\n", totalLines)
		}

		flushBatch := func() error {
			if len(batch) == 0 {
				return nil
			}

			req := scim.BulkRequest{
				Operations: batch,
			}

			resp, err := client.Bulk(ctx, req)
			if err != nil {
				return fmt.Errorf("bulk request failed: %w", err)
			}

			for _, opResp := range resp.Operations {
				if opResp.Status == 201 {
					totalSuccess++
				} else {
					totalErrors++
				}
				// Print the result
				jsonData, err := json.Marshal(opResp)
				if err != nil {
					fmt.Fprintln(os.Stderr, "\nFailed to marshal response:", err)
				} else {
					fmt.Println(string(jsonData))
				}
			}

			// Clear the batch
			batch = batch[:0]
			return nil
		}

		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var resourceData scim.Resource
			if err := json.Unmarshal([]byte(line), &resourceData); err != nil {
				fmt.Fprintf(os.Stderr, "\nError parsing JSON at line %d: %v\n", lineNum, err)
				totalErrors++
				continue
			}

			op := scim.BulkOperation{
				Method: "POST",
				Path:   resourcePath,
				BulkId: uuid.New().String(),
				Data:   resourceData,
			}

			batch = append(batch, op)

			if len(batch) >= importChunkSize {
				if err := flushBatch(); err != nil {
					return err
				}
				
				processedCount += importChunkSize
				
				if isStderrTerminal {
					if totalLines > 0 {
						currentPercent := (processedCount * 100) / totalLines
						if currentPercent > lastReportedPercent {
							fmt.Fprintf(os.Stderr, "Importing... %d%%\r", currentPercent)
							lastReportedPercent = currentPercent
						}
					} else {
						fmt.Fprintf(os.Stderr, "Importing... %d processed\r", processedCount)
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading input stream: %w", err)
		}

		// Flush remaining
		remainingCount := len(batch)
		if err := flushBatch(); err != nil {
			return err
		}
		
		processedCount += remainingCount
		if isStderrTerminal {
			if totalLines > 0 {
				fmt.Fprintf(os.Stderr, "Importing... 100%%\n")
			} else {
				fmt.Fprintf(os.Stderr, "Importing... %d processed\n", processedCount)
			}
		}

		report := map[string]interface{}{
			"message":      "Import complete",
			"totalSuccess": totalSuccess,
			"totalErrors":  totalErrors,
		}

		if reportJSON, err := json.Marshal(report); err == nil {
			fmt.Println(string(reportJSON))
		} else {
			fmt.Fprintf(os.Stderr, "Failed to marshal final report: %v\n", err)
		}

		if totalErrors > 0 {
			return fmt.Errorf("import completed with %d errors", totalErrors)
		}

		return nil
	},
}

// Note: countLines function will be shared from bulk_update.go in the same package

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringVarP(&importResourceType, "resource", "r", "", "SCIM resource type (required)")
	importCmd.Flags().StringVarP(&importFile, "file", "f", "", "Input JSON Lines file path")
	importCmd.Flags().IntVar(&importChunkSize, "chunk", 100, "Chunk size for bulk requests")
	importCmd.MarkFlagRequired("resource")
}
