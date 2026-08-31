package commands

import (
	"bufio"
	"bytes"
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
	bulkUpdateResourceType string
	bulkUpdateFile         string
	bulkUpdateChunkSize    int
)

type bulkUpdatePayload struct {
	ID         string                `json:"id"`
	Operations []scim.PatchOperation `json:"operations"`
}

// bulkUpdateCmd represents the bulk-update command
var bulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Update SCIM resources using the Bulk API",
	Long: `Update SCIM resources on the server using the Bulk API (PATCH operations).
The input data should be a stream of JSON Lines.
Each line must contain the resource "id" and an array of "operations".
The data can be provided via the --file flag or through STDIN.

Examples:
  scim-ctl bulk-update --resource user --file updates.jsonl
  cat updates.jsonl | scim-ctl bulk-update -r user --chunk 50`,
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

		if bulkUpdateFile != "" {
			// Count total lines for progress tracking
			file, err := os.Open(bulkUpdateFile)
			if err != nil {
				return fmt.Errorf("failed to open input file: %w", err)
			}
			totalLines = countLines(file)
			file.Close()

			// Reopen for actual reading
			file, err = os.Open(bulkUpdateFile)
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

		resourcePathPrefix := "/" + scim.ResourceName(bulkUpdateResourceType)

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
				// Status 200 is OK for PATCH in Bulk
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
			fmt.Fprintf(os.Stderr, "Total resources to update: %d\n", totalLines)
		}

		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var payload bulkUpdatePayload
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

			if len(payload.Operations) == 0 {
				fmt.Fprintf(os.Stderr, "\nError at line %d: missing 'operations'\n", lineNum)
				totalErrors++
				continue
			}

			// Construct PATCH data payload
			patchData := map[string]interface{}{
				"schemas":    []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
				"Operations": payload.Operations,
			}

			op := scim.BulkOperation{
				Method: "PATCH",
				Path:   resourcePathPrefix + "/" + payload.ID,
				BulkId: uuid.New().String(),
				Data:   patchData,
			}

			batch = append(batch, op)

			if len(batch) >= bulkUpdateChunkSize {
				if err := flushBatch(); err != nil {
					return err
				}
				
				processedCount += bulkUpdateChunkSize
				
				if isStderrTerminal {
					if totalLines > 0 {
						currentPercent := (processedCount * 100) / totalLines
						if currentPercent > lastReportedPercent {
							fmt.Fprintf(os.Stderr, "Updating... %d%%\r", currentPercent)
							lastReportedPercent = currentPercent
						}
					} else {
						fmt.Fprintf(os.Stderr, "Updating... %d processed\r", processedCount)
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
				fmt.Fprintf(os.Stderr, "Updating... 100%%\n")
			} else {
				fmt.Fprintf(os.Stderr, "Updating... %d processed\n", processedCount)
			}
		}

		report := map[string]interface{}{
			"message":      "Bulk update complete",
			"totalSuccess": totalSuccess,
			"totalErrors":  totalErrors,
		}

		if reportJSON, err := json.Marshal(report); err == nil {
			fmt.Println(string(reportJSON))
		} else {
			fmt.Fprintf(os.Stderr, "Failed to marshal final report: %v\n", err)
		}

		if totalErrors > 0 {
			return fmt.Errorf("bulk update completed with %d errors", totalErrors)
		}

		return nil
	},
}

func countLines(r io.Reader) int {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count
		case err != nil:
			return count // best effort
		}
	}
}

func init() {
	rootCmd.AddCommand(bulkUpdateCmd)

	bulkUpdateCmd.Flags().StringVarP(&bulkUpdateResourceType, "resource", "r", "", "SCIM resource type (required)")
	bulkUpdateCmd.Flags().StringVarP(&bulkUpdateFile, "file", "f", "", "Input JSON Lines file path")
	bulkUpdateCmd.Flags().IntVar(&bulkUpdateChunkSize, "chunk", 100, "Chunk size for bulk requests")
	bulkUpdateCmd.MarkFlagRequired("resource")
}
