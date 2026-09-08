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
	bulkDeleteResourceType string
	bulkDeleteFile         string
	bulkDeleteChunkSize    int
)

type bulkDeletePayload struct {
	ID string `json:"id"`
}

// bulkDeleteCmd represents the bulk-delete command
var bulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete",
	Short: "Delete SCIM resources using the Bulk API",
	Long: `Delete SCIM resources on the server using the Bulk API (DELETE operations).
The input data should be a stream of JSON Lines.
Each line must contain the resource "id".
The data can be provided via the --file flag or through STDIN.

Examples:
  scim-ctl bulk-delete --resource user --file deletes.jsonl
  cat deletes.jsonl | scim-ctl bulk-delete -r user --chunk 50`,
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

		var reader io.Reader
		totalLines := 0

		if bulkDeleteFile != "" {
			// Count total lines for progress tracking
			file, err := os.Open(bulkDeleteFile)
			if err != nil {
				return fmt.Errorf("failed to open input file: %w", err)
			}
			totalLines = countLines(file)
			file.Close()

			// Reopen for actual reading
			file, err = os.Open(bulkDeleteFile)
			if err != nil {
				return fmt.Errorf("failed to reopen input file: %w", err)
			}
			defer file.Close()
			reader = file
		} else {
			stat, err := os.Stdin.Stat()
			if err != nil {
				return fmt.Errorf("failed to check STDIN: %w", err)
			}
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return fmt.Errorf("input data is required (use --file or pipe via STDIN)")
			}
			reader = os.Stdin
		}

		stderrStat, _ := os.Stderr.Stat()
		isStderrTerminal := (stderrStat.Mode() & os.ModeCharDevice) != 0

		scanner := bufio.NewScanner(reader)
		const maxCapacity = 10 * 1024 * 1024 // 10MB
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		var batch []scim.BulkOperation
		var totalSuccess, totalErrors int
		processedCount := 0
		lastReportedPercent := -1

		resourcePathPrefix := "/" + scim.ResourceName(bulkDeleteResourceType)

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
				// Status 200 or 204 is OK for DELETE in Bulk
				if opResp.Status == 200 || opResp.Status == 204 {
					totalSuccess++
				} else {
					totalErrors++
				}
				// Print the result to stdout
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

		if isStderrTerminal && totalLines > 0 {
			fmt.Fprintf(os.Stderr, "Total resources to delete: %d\n", totalLines)
		}

		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var payload bulkDeletePayload
			if err := json.Unmarshal([]byte(line), &payload); err != nil {
				fmt.Fprintf(os.Stderr, "\nError parsing JSON at line %d: %v\n", lineNum, err)
				totalErrors++
				continue
			}

			if payload.ID == "" {
				fmt.Fprintf(os.Stderr, "\nError at line %d: missing resource 'id'\n", lineNum)
				totalErrors++
				continue
			}

			op := scim.BulkOperation{
				Method: "DELETE",
				Path:   resourcePathPrefix + "/" + payload.ID,
				BulkId: uuid.New().String(),
			}

			batch = append(batch, op)

			if len(batch) >= bulkDeleteChunkSize {
				if err := flushBatch(); err != nil {
					return err
				}
				
				processedCount += bulkDeleteChunkSize
				
				if isStderrTerminal {
					if totalLines > 0 {
						currentPercent := (processedCount * 100) / totalLines
						if currentPercent > lastReportedPercent {
							fmt.Fprintf(os.Stderr, "Deleting... %d%%\r", currentPercent)
							lastReportedPercent = currentPercent
						}
					} else {
						fmt.Fprintf(os.Stderr, "Deleting... %d processed\r", processedCount)
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
				fmt.Fprintf(os.Stderr, "Deleting... 100%%\n")
			} else {
				fmt.Fprintf(os.Stderr, "Deleting... %d processed\n", processedCount)
			}
		}

		report := map[string]interface{}{
			"message":      "Bulk delete complete",
			"totalSuccess": totalSuccess,
			"totalErrors":  totalErrors,
		}

		if reportJSON, err := json.Marshal(report); err == nil {
			fmt.Println(string(reportJSON))
		} else {
			fmt.Fprintf(os.Stderr, "Failed to marshal final report: %v\n", err)
		}

		if totalErrors > 0 {
			return fmt.Errorf("bulk delete completed with %d errors", totalErrors)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(bulkDeleteCmd)

	bulkDeleteCmd.Flags().StringVarP(&bulkDeleteResourceType, "resource", "r", "", "SCIM resource type (required)")
	bulkDeleteCmd.Flags().StringVarP(&bulkDeleteFile, "file", "f", "", "Input JSON Lines file path")
	bulkDeleteCmd.Flags().IntVar(&bulkDeleteChunkSize, "chunk", 100, "Chunk size for bulk requests")
	bulkDeleteCmd.MarkFlagRequired("resource")
}
