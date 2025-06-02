package cmd

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/regiellis/chronos-go/db"
	"github.com/regiellis/chronos-go/ui"
	"github.com/regiellis/chronos-go/utils"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "View entries, blocks, and summaries",
}

var viewBlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Show the active block and its progress",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		if err := dbStore.InitSchema(); err != nil {
			return err
		}
		block, err := dbStore.GetActiveBlock()
		if err != nil {
			return err
		}
		if block == nil {
			fmt.Println(utils.ErrorStyle.Render("No active block."))
			return nil
		}
		fmt.Println(utils.TitleStyle.Render("Active Block"))
		fmt.Println(utils.EntryStyle.Render(
			fmt.Sprintf("Name: %s\nClient: %s\nProject: %s\nStart: %s\nEnd: %s",
				utils.SanitizeString(block.Name),
				utils.SanitizeString(block.Client),
				utils.SanitizeString(block.Project),
				block.StartTime.Format("2006-01-02"),
				block.EndTime.Format("2006-01-02"),
			),
		))
		return nil
	},
}

var (
	filterBlockID  int64
	filterProject  string
	filterClient   string
	filterTask     string
	filterFrom     string
	filterTo       string
	filterMinDur   int64
	filterMaxDur   int64
	filterBillable bool
	filterMinRate  float64
	filterMaxRate  float64
)

var viewListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show a list of time entries (filterable)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		if err := dbStore.InitSchema(); err != nil {
			return err
		}
		filter := map[string]interface{}{}
		if filterBlockID > 0 {
			filter["block_id"] = filterBlockID
		}
		if filterProject != "" {
			filter["project"] = filterProject
		}
		if filterClient != "" {
			filter["client"] = filterClient
		}
		if filterTask != "" {
			filter["task"] = filterTask
		}
		if filterFrom != "" {
			if t, err := time.Parse("2006-01-02", filterFrom); err == nil {
				filter["from"] = t.Format(time.RFC3339)
			}
		}
		if filterTo != "" {
			if t, err := time.Parse("2006-01-02", filterTo); err == nil {
				filter["to"] = t.Format(time.RFC3339)
			}
		}
		if filterMinDur > 0 {
			filter["min_duration"] = filterMinDur
		}
		if filterMaxDur > 0 {
			filter["max_duration"] = filterMaxDur
		}
		if filterBillable {
			filter["billable"] = true
		}
		if filterMinRate > 0 {
			filter["min_rate"] = filterMinRate
		}
		if filterMaxRate > 0 {
			filter["max_rate"] = filterMaxRate
		}
		entries, err := dbStore.ListEntries(filter)
		if err != nil {
			return err
		}
		model := ui.NewListViewModel(entries)
		p := tea.NewProgram(model)
		return p.Start()
	},
}

var viewInvoiceCmd = &cobra.Command{
	Use:   "invoice",
	Short: "Show invoice-ready summary for a block or client",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		if err := dbStore.InitSchema(); err != nil {
			return err
		}
		blockID, _ := cmd.Flags().GetInt64("block")
		client, _ := cmd.Flags().GetString("client")
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
		fmt.Println(utils.TitleStyle.Render("Invoice Summary"))
		fmt.Println(utils.EntryStyle.Render(
			fmt.Sprintf("Billable entries: %d\nTotal hours: %.2f\nTotal amount: $%.2f",
				len(entries), float64(totalMinutes)/60.0, totalAmount,
			),
		))
		return nil
	},
}

var invoiceMDViewCmd = &cobra.Command{
	Use:   "invoice-md",
	Short: "Render invoice as Markdown in the terminal (with Glamour)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		blockID, _ := cmd.Flags().GetInt64("block")
		client, _ := cmd.Flags().GetString("client")
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
			hours := float64(e.Duration) / 60.0
			totalMinutes += e.Duration
			totalAmount += hours * e.Rate
		}
		md := "# Invoice\n\n| Project | Task | Description | Hours | Rate | Amount |\n|---|---|---|---|---|---|\n"
		for _, e := range entries {
			hours := float64(e.Duration) / 60.0
			amt := hours * e.Rate
			md += fmt.Sprintf("| %s | %s | %s | %.2f | %.2f | %.2f |\n", e.Project, e.Task, e.Description, hours, e.Rate, amt)
		}
		md += fmt.Sprintf("\n**Total Hours:** %.2f\n**Total Amount:** $%.2f\n", float64(totalMinutes)/60.0, totalAmount)
		fmt.Println(utils.TitleStyle.Render("Invoice (Markdown Preview)"))
		fmt.Println(md)
		return nil
	},
}

func init() {
	viewListCmd.Flags().Int64Var(&filterBlockID, "block", 0, "Filter by block ID")
	viewListCmd.Flags().StringVar(&filterProject, "project", "", "Filter by project")
	viewListCmd.Flags().StringVar(&filterClient, "client", "", "Filter by client")
	viewListCmd.Flags().StringVar(&filterTask, "task", "", "Filter by task")
	viewListCmd.Flags().StringVar(&filterFrom, "from", "", "Filter from date (YYYY-MM-DD)")
	viewListCmd.Flags().StringVar(&filterTo, "to", "", "Filter to date (YYYY-MM-DD)")
	viewListCmd.Flags().Int64Var(&filterMinDur, "min-duration", 0, "Filter by minimum duration (minutes)")
	viewListCmd.Flags().Int64Var(&filterMaxDur, "max-duration", 0, "Filter by maximum duration (minutes)")
	viewListCmd.Flags().BoolVar(&filterBillable, "billable", false, "Show only billable entries")
	viewListCmd.Flags().Float64Var(&filterMinRate, "min-rate", 0, "Filter by minimum rate")
	viewListCmd.Flags().Float64Var(&filterMaxRate, "max-rate", 0, "Filter by maximum rate")

	viewInvoiceCmd.Flags().Int64("block", 0, "Block ID to invoice")
	viewInvoiceCmd.Flags().String("client", "", "Client to invoice")
	viewCmd.AddCommand(viewInvoiceCmd)

	invoiceMDViewCmd.Flags().Int64("block", 0, "Block ID to invoice")
	invoiceMDViewCmd.Flags().String("client", "", "Client to invoice")
	viewCmd.AddCommand(invoiceMDViewCmd)

	viewCmd.AddCommand(viewBlockCmd)
	viewCmd.AddCommand(viewListCmd)
	rootCmd.AddCommand(viewCmd)
}
