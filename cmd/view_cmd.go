package cmd

import (
	"database/sql"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/regiellis/chronos-go/chronos" // Imported chronos
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
			return fmt.Errorf("failed to create db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("failed to initialize schema: %w", err)
		}

		// Use new chronos function
		activeBlock, err := chronos.GetActiveBlock(dbStore)
		if err != nil {
			// sql.ErrNoRows is a valid case where no block is active, not necessarily a fatal error.
			if err == sql.ErrNoRows {
				fmt.Println(utils.InfoStyle.Render("No active block."))
				return nil
			}
			return fmt.Errorf("failed to get active block: %w", err)
		}
		if activeBlock == nil { // Should be covered by sql.ErrNoRows, but good for safety
			fmt.Println(utils.InfoStyle.Render("No active block."))
			return nil
		}

		fmt.Println(utils.TitleStyle.Render("Active Block"))
		// Assuming chronos.Block struct fields are compatible with this display.
		// chronos.Block has Name, Client (string), Project (string), StartTime, EndTime.
		fmt.Println(utils.EntryStyle.Render(
			fmt.Sprintf("ID: %d\nName: %s\nClient: %s\nProject: %s\nStart: %s\nEnd: %s\nActive: %t",
				activeBlock.ID, // Added ID
				utils.SanitizeString(activeBlock.Name),
				utils.SanitizeString(activeBlock.Client),   // Assumes Client is a string name
				utils.SanitizeString(activeBlock.Project), // Assumes Project is a string name
				activeBlock.StartTime.Format("2006-01-02 15:04"), // More precise time
				activeBlock.EndTime.Format("2006-01-02 15:04"),   // More precise time
				activeBlock.Active, // Added Active status
			),
		))
		return nil
	},
}

var (
	filterBlockID  int64
	filterProject  string // Project name
	filterClient   string // Client name
	// filterTask     string // Task filtering is not directly supported by new ListEntries; would require summary LIKE '%task%'
	filterFrom     string
	filterTo       string
	// filterMinDur   int64 // Duration filters not directly supported; require calculation from Start/End
	// filterMaxDur   int64
	filterBillable bool // Maps to "invoiced"
	// filterMinRate  float64 // Rate is not on chronos.Entry; requires joining with Project or other logic
	// filterMaxRate  float64
)

var viewListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show a list of time entries (filterable)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("failed to initialize schema: %w", err)
		}

		chronosFilters := map[string]interface{}{}
		if filterBlockID > 0 {
			chronosFilters["block_id"] = filterBlockID
		}
		if filterBillable { // Assuming "billable" flag maps to "invoiced" status
			chronosFilters["invoiced"] = true // Or false, depending on desired logic if flag could mean "only unbilled"
		}
		if filterFrom != "" {
			if t, errDate := time.Parse("2006-01-02", filterFrom); errDate == nil {
				chronosFilters["start_date"] = t // chronos.ListEntries expects time.Time for these
			} else {
				fmt.Printf("Warning: could not parse 'from' date '%s': %v\n", filterFrom, errDate)
			}
		}
		if filterTo != "" {
			if t, errDate := time.Parse("2006-01-02", filterTo); errDate == nil {
				chronosFilters["end_date"] = t // chronos.ListEntries expects time.Time
			} else {
				fmt.Printf("Warning: could not parse 'to' date '%s': %v\n", filterTo, errDate)
			}
		}
		
		// TODO: Filtering by Project Name (filterProject) requires a lookup:
		// if filterProject != "" {
		//   project, err := chronos.GetProjectByName(dbStore, filterProject) // Hypothetical function
		//   if err == nil && project != nil {
		//     chronosFilters["project_id"] = project.ID
		//   } else {
		//     fmt.Printf("Warning: Project '%s' not found, filter not applied.\n", filterProject)
		//   }
		// }
		// Client, Task, Duration, and Rate filters are not directly supported by the new chronos.ListEntries.
		// These would require more complex logic (e.g., joining tables, post-filtering, or enhancing ListEntries).

		entries, err := chronos.ListEntries(dbStore, chronosFilters)
		if err != nil {
			return fmt.Errorf("failed to list entries: %w", err)
		}

		// Assuming ui.NewListViewModel is compatible with []*chronos.Entry
		model := ui.NewListViewModel(entries)
		p := tea.NewProgram(model)
		return p.Start()
	},
}

// Note: The invoice commands below will be significantly impacted by chronos.Entry struct changes.
// chronos.Entry (ID, BlockID, ProjectID, Summary, StartTime, EndTime, CreatedAt, UpdatedAt, Invoiced)
// does not have direct fields for Rate, Duration (must be calculated), Project (name), Task (name).
// These commands will need substantial rework to fetch related data (e.g., Project details for rate)
// or the chronos.Entry struct/ListEntries function needs to be enhanced.

var viewInvoiceCmd = &cobra.Command{
	Use:   "invoice",
	Short: "Show invoice-ready summary (functionality limited post-refactor)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil { return fmt.Errorf("db store: %w", err) }
		if err := dbStore.InitSchema(); err != nil { return fmt.Errorf("schema init: %w", err) }

		blockID, _ := cmd.Flags().GetInt64("block")
		// clientName, _ := cmd.Flags().GetString("client") // Client name filtering is complex now

		chronosFilters := map[string]interface{}{"invoiced": false} // Typically invoice un-invoiced entries
		if blockID > 0 {
			chronosFilters["block_id"] = blockID
		}
		// TODO: Add project filtering if needed, via project ID lookup.

		entries, err := chronos.ListEntries(dbStore, chronosFilters)
		if err != nil { return fmt.Errorf("list entries: %w", err) }

		var totalMinutes float64
		var totalAmount float64
		fmt.Println(utils.WarningStyle.Render("Warning: Invoice calculation may be inaccurate due to missing Rate and direct Duration fields in refactored Entry struct."))

		for _, e := range entries {
			duration := e.EndTime.Sub(e.StartTime).Minutes()
			totalMinutes += duration
			// TODO: Need to fetch Project Rate for accurate billing. Using placeholder rate 0.
			// project, _ := chronos.GetProjectByID(dbStore, e.ProjectID)
			// rate := project.Rate (if project is not nil)
			rate := 0.0 // Placeholder
			totalAmount += (duration / 60.0) * rate
		}

		fmt.Println(utils.TitleStyle.Render("Invoice Summary (Limited)"))
		fmt.Println(utils.EntryStyle.Render(
			fmt.Sprintf("Billable entries: %d\nTotal hours: %.2f\nTotal amount: $%.2f (Rate calculation needs rework)",
				len(entries), totalMinutes/60.0, totalAmount,
			),
		))
		return nil
	},
}

var invoiceMDViewCmd = &cobra.Command{
	Use:   "invoice-md",
	Short: "Render invoice as Markdown (functionality limited post-refactor)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil { return fmt.Errorf("db store: %w", err) }
		if err := dbStore.InitSchema(); err != nil { return fmt.Errorf("schema init: %w", err) }

		blockID, _ := cmd.Flags().GetInt64("block")
		// clientName, _ := cmd.Flags().GetString("client")

		chronosFilters := map[string]interface{}{"invoiced": false}
		if blockID > 0 {
			chronosFilters["block_id"] = blockID
		}

		entries, err := chronos.ListEntries(dbStore, chronosFilters)
		if err != nil { return fmt.Errorf("list entries: %w", err) }
		
		fmt.Println(utils.WarningStyle.Render("Warning: Invoice Markdown generation may be inaccurate or incomplete."))

		var totalMinutes float64
		var totalAmount float64
		md := "# Invoice (Limited)\n\n| Project ID | Summary | Hours | Rate | Amount |\n|---|---|---|---|---|\n"
		for _, e := range entries {
			hours := e.EndTime.Sub(e.StartTime).Minutes() / 60.0
			totalMinutes += hours * 60
			// TODO: Fetch Project & Rate. Using placeholders.
			// project, _ := chronos.GetProjectByID(dbStore, e.ProjectID)
			// projectName := "N/A" (if project is nil else project.Name)
			// rate := 0.0 (if project is nil else project.Rate)
			projectName := fmt.Sprintf("ProjID %d", e.ProjectID)
			rate := 0.0 // Placeholder
			amt := hours * rate
			totalAmount += amt
			// Summary is the main text. Task/Description separation is gone from chronos.Entry.
			md += fmt.Sprintf("| %s | %s | %.2f | %.2f | %.2f |\n", projectName, e.Summary, hours, rate, amt)
		}
		md += fmt.Sprintf("\n**Total Hours:** %.2f\n**Total Amount:** $%.2f (Rate calculation needs rework)\n", totalMinutes/60.0, totalAmount)
		
		fmt.Println(utils.TitleStyle.Render("Invoice (Markdown Preview - Limited)"))
		fmt.Println(md) // In a real scenario, this would go through Glamour or similar.
		return nil
	},
}


func init() {
	// Flags for viewListCmd
	viewListCmd.Flags().Int64Var(&filterBlockID, "block", 0, "Filter by block ID")
	viewListCmd.Flags().StringVar(&filterProject, "project", "", "Filter by project name (requires lookup)")
	viewListCmd.Flags().StringVar(&filterClient, "client", "", "Filter by client name (currently not supported)")
	// viewListCmd.Flags().StringVar(&filterTask, "task", "", "Filter by task (currently not supported)")
	viewListCmd.Flags().StringVar(&filterFrom, "from", "", "Filter from date (YYYY-MM-DD)")
	viewListCmd.Flags().StringVar(&filterTo, "to", "", "Filter to date (YYYY-MM-DD)")
	// viewListCmd.Flags().Int64Var(&filterMinDur, "min-duration", 0, "Filter by minimum duration (minutes, currently not supported)")
	// viewListCmd.Flags().Int64Var(&filterMaxDur, "max-duration", 0, "Filter by maximum duration (minutes, currently not supported)")
	viewListCmd.Flags().BoolVar(&filterBillable, "billable", false, "Show only 'billable' (invoiced=true or invoiced=false based on interpretation) entries")
	// viewListCmd.Flags().Float64Var(&filterMinRate, "min-rate", 0, "Filter by minimum rate (currently not supported)")
	// viewListCmd.Flags().Float64Var(&filterMaxRate, "max-rate", 0, "Filter by maximum rate (currently not supported)")


	viewInvoiceCmd.Flags().Int64("block", 0, "Block ID to invoice")
	viewInvoiceCmd.Flags().String("client", "", "Client to invoice (currently not directly supported for filtering)")
	viewCmd.AddCommand(viewInvoiceCmd)

	invoiceMDViewCmd.Flags().Int64("block", 0, "Block ID to invoice")
	invoiceMDViewCmd.Flags().String("client", "", "Client to invoice (currently not directly supported for filtering)")
	viewCmd.AddCommand(invoiceMDViewCmd)

	viewCmd.AddCommand(viewBlockCmd)
	viewCmd.AddCommand(viewListCmd)
	rootCmd.AddCommand(viewCmd)
}
