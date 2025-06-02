package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/regiellis/chronos-go/chronos"
	"github.com/regiellis/chronos-go/db"
	"github.com/regiellis/chronos-go/utils"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data as JSON",
}

var exportSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Export summary as JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		if err := dbStore.InitSchema(); err != nil {
			return err
		}
		entries, err := dbStore.ListEntries(nil)
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(utils.SuccessStyle.Render("Exported summary as JSON:"))
		fmt.Println(string(data))
		return nil
	},
}

var exportSuggestCmd = &cobra.Command{
	Use:   "suggestion",
	Short: "Export the latest LLM suggestion or answer as JSON or Markdown",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(utils.SuccessStyle.Render("Exporting last suggestion/answer as JSON/Markdown (feature stub)"))
		return nil
	},
}

var exportInvoiceCmd = &cobra.Command{
	Use:   "invoice",
	Short: "Export invoice view as JSON or Markdown",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		blockID, _ := cmd.Flags().GetInt64("block")
		client, _ := cmd.Flags().GetString("client")
		format, _ := cmd.Flags().GetString("format")
		filter := map[string]interface{}{"billable": true}
		if blockID > 0 {
			filter["block_id"] = blockID
		}
		if client != "" {
			filter["client"] = client
		}
		entries, err := dbStore.ListEntries(filter)
		if err != nil {
			return err
		}
		var totalMinutes int64
		var totalAmount float64
		for _, e := range entries {
			totalMinutes += e.Duration
			totalAmount += (float64(e.Duration) / 60.0) * e.Rate
		}
		invoice := struct {
			Entries     []*chronos.Entry `json:"entries"`
			TotalHours  float64          `json:"total_hours"`
			TotalAmount float64          `json:"total_amount"`
		}{
			Entries:     entries,
			TotalHours:  float64(totalMinutes) / 60.0,
			TotalAmount: totalAmount,
		}
		if format == "markdown" {
			fmt.Println(utils.TitleStyle.Render("Invoice (Markdown Export)"))
			fmt.Println("# Invoice\n\n| Project | Task | Description | Hours | Rate | Amount |\n|---|---|---|---|---|---|")
			for _, e := range entries {
				hours := float64(e.Duration) / 60.0
				amt := hours * e.Rate
				fmt.Println(fmt.Sprintf("| %s | %s | %s | %.2f | %.2f | %.2f |", e.Project, e.Task, e.Description, hours, e.Rate, amt))
			}
			fmt.Println(fmt.Sprintf("\n**Total Hours:** %.2f\n**Total Amount:** $%.2f\n", invoice.TotalHours, invoice.TotalAmount))
			return nil
		}
		data, err := json.MarshalIndent(invoice, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(utils.SuccessStyle.Render("Exported invoice as JSON:"))
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	exportCmd.AddCommand(exportSummaryCmd)
	exportCmd.AddCommand(exportSuggestCmd)
	exportInvoiceCmd.Flags().Int64("block", 0, "Block ID to invoice")
	exportInvoiceCmd.Flags().String("client", "", "Client to invoice")
	exportInvoiceCmd.Flags().String("format", "json", "Export format: json or markdown")
	exportCmd.AddCommand(exportInvoiceCmd)
	rootCmd.AddCommand(exportCmd)
}
