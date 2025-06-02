/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"
	"github.com/regiellis/chronos-go/chronos" // Imported chronos
	"github.com/regiellis/chronos-go/config"
	"github.com/regiellis/chronos-go/db"
	"github.com/regiellis/chronos-go/llm"
	"github.com/regiellis/chronos-go/ui"
	"github.com/regiellis/chronos-go/utils"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chronos",
	Short: utils.TitleStyle.Render("Chronos: AI-powered time tracking and productivity CLI"),
	Long: lipgloss.JoinVertical(lipgloss.Left,
		utils.TitleStyle.Render("Chronos: AI-powered Time Tracking & Productivity"),
		utils.SubtitleStyle.Render("A beautiful, intelligent CLI for freelancers and makers"),
		"",
		utils.LabelStyle.Render("Features include:"),
		utils.ValueStyle.Render("  • Fast time entry logging and editing"),
		utils.ValueStyle.Render("  • AI-powered suggestions, reminders, and auto-completion"),
		utils.ValueStyle.Render("  • Natural language queries about your tracked time and progress"),
		utils.ValueStyle.Render("  • Pomodoro/focus timer with logging"),
		utils.ValueStyle.Render("  • Idle detection and smart nudges"),
		utils.ValueStyle.Render("  • Analytics for clients, projects, and tasks"),
		utils.ValueStyle.Render("  • Entry templates and snippets"),
		utils.ValueStyle.Render("  • Invoice and billing support"),
		utils.ValueStyle.Render("  • Weekly/monthly reviews with LLM summaries"),
		"",
		utils.InactiveStyle.Render("See 'chronos help [command]' for details on each feature."),
	),
}

// askCmd represents the ask command
var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask the LLM about your tracked time, blocks, or progress",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("schema init: %w", err)
		}
		llmClient := llm.NewOllamaClient()
		question := strings.Join(args, " ")
		entries, _ := chronos.ListEntries(dbStore, nil) // Refactored
		blocks, _ := chronos.ListBlocks(dbStore, nil)   // Refactored
		answer, err := llmClient.AnswerUserQuery(question, entries, blocks)
		if err != nil {
			return err
		}
		log.Info(answer)
		return nil
	},
}

// suggestCmd represents the suggest command
var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Get a suggestion for your next entry or task",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("schema init: %w", err)
		}
		llmClient := llm.NewOllamaClient()
		entries, _ := chronos.ListEntries(dbStore, nil) // Refactored
		blocks, _ := chronos.ListBlocks(dbStore, nil)   // Refactored
		suggestion, err := llmClient.SuggestNextEntry(entries, blocks)
		if err != nil {
			return err
		}
		log.Info(suggestion)
		return nil
	},
}

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query [question]",
	Short: "Ask a natural language question (shortcut for 'ask')",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return askCmd.RunE(cmd, args) // Already uses askCmd which is refactored
	},
}

// remindCmd represents the remind command
var remindCmd = &cobra.Command{
	Use:   "remind",
	Short: "Show smart reminders or nudges based on your activity",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("schema init: %w", err)
		}
		llmClient := llm.NewOllamaClient()
		entries, _ := chronos.ListEntries(dbStore, nil) // Refactored
		blocks, _ := chronos.ListBlocks(dbStore, nil)   // Refactored
		reminder, err := llmClient.SmartReminder(entries, blocks)
		if err != nil {
			return err
		}
		log.Info(reminder)
		return nil
	},
}

// historyCmd represents the history command (manages LLM query history, not chronos entries/blocks)
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show your recent LLM queries for quick re-use",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath) // This is a direct db.Store usage for its specific QueryHistory
		if err != nil {
			return err
		}
		// QueryHistory is specific to db.Store and not part of chronos package's concerns for now
		queries, err := dbStore.QueryHistory(10)
		if err != nil {
			return err
		}
		for i, q := range queries {
			log.Info("%d: %s", i+1, q)
		}
		return nil
	},
}

// completeCmd represents the complete command
var completeCmd = &cobra.Command{
	Use:   "complete [partial]",
	Short: "Get LLM-powered auto-completions for project, client, or task fields",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil { // Ensure schema for List functions
			return fmt.Errorf("schema init: %w", err)
		}
		llmClient := llm.NewOllamaClient()
		entries, _ := chronos.ListEntries(dbStore, nil) // Refactored
		blocks, _ := chronos.ListBlocks(dbStore, nil)   // Refactored
		partial := args[0]
		suggestion, err := llmClient.AutoCompleteFields(partial, entries, blocks)
		if err != nil {
			return err
		}
		log.Info(suggestion)
		return nil
	},
}

var editCmd = &cobra.Command{
	Use:   "edit [entry_id]",
	Short: "Edit a time entry by ID (toggles invoiced status for now)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("schema init: %w", err)
		}
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid entry ID: %w", err)
		}

		entry, err := chronos.GetEntryByID(dbStore, id) // Refactored
		if err != nil {
			if err == sql.ErrNoRows || strings.Contains(err.Error(), "no entry found") { // Check for custom error from GetEntryByID
				log.Error("Entry not found", "ID", id)
			} else {
				log.Error("Failed to get entry", "ID", id, "error", err)
			}
			return fmt.Errorf("could not retrieve entry %d: %w", id, err)
		}

		// For demo: just toggle invoiced status
		entry.Invoiced = !entry.Invoiced
		entry.UpdatedAt = time.Now() // Update timestamp

		if err := chronos.UpdateEntry(dbStore, entry); err != nil { // Refactored
			log.Error("Failed to update entry", "ID", id, "error", err)
			return fmt.Errorf("could not update entry %d: %w", id, err)
		}
		log.Info("Entry updated successfully", "ID", id, "Invoiced", entry.Invoiced)
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [entry_id]",
	Short: "Delete a time entry by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil { // Ensure schema for DeleteEntry
			return fmt.Errorf("schema init: %w", err)
		}
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid entry ID: %w", err)
		}

		if err := chronos.DeleteEntry(dbStore, id); err != nil { // Refactored
			log.Error("Failed to delete entry", "ID", id, "error", err)
			return fmt.Errorf("could not delete entry %d: %w", id, err)
		}
		log.Info("Entry deleted successfully", "ID", id)
		return nil
	},
}

var pomodoroCmd = &cobra.Command{
	Use:   "pomodoro [duration]",
	Short: "Start a Pomodoro/focus session and log it as an entry",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dur := 25 * time.Minute
		if len(args) > 0 {
			if d, err := time.ParseDuration(args[0]); err == nil {
				dur = d
			}
		}
		// Pomodoro UI and logic for creating entry after completion is in ui.PomodoroModel.
		// This part doesn't directly use chronos CRUD yet, but ui.PomodoroModel might internally.
		// If ui.PomodoroModel needs db.Store, it should be passed.
		// For now, assuming it handles its own DB interaction or this will be refactored later.
		model := ui.NewPomodoroModel(dur) // This might need dbStore if it creates an entry.
		p := tea.NewProgram(model)
		return p.Start()
	},
}

var idleCmd = &cobra.Command{
	Use:   "idle-detect",
	Short: "Detect idle gaps and suggest logging missed time",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("schema init: %w", err)
		}
		// Fetch all entries. DetectIdleGaps will sort them.
		entries, err := chronos.ListEntries(dbStore, nil) 
		if err != nil {
			return fmt.Errorf("failed to list entries for idle detection: %w", err)
		}

		if len(entries) < 2 {
			log.Warn("Not enough entries to detect idle gaps.")
			return nil
		}

		minGapDuration := 2 * time.Hour // Define the minimum duration to consider as an idle gap
		idleGaps := chronos.DetectIdleGaps(entries, minGapDuration)

		if len(idleGaps) == 0 {
			log.Info("No significant idle gaps detected.")
			return nil
		}

		log.Info("Detected Idle Gaps (longer than %v):", minGapDuration)
		for _, gap := range idleGaps {
			log.Warn(fmt.Sprintf("Gap from %s to %s (Duration: %v)",
				gap.StartTime.Format("2006-01-02 15:04"),
				gap.EndTime.Format("2006-01-02 15:04"),
				gap.Duration.Round(time.Minute), // Rounded for cleaner output
			))
		}
		return nil
	},
}

var rateCmd = &cobra.Command{
	Use:   "rate [amount]",
	Short: "Quickly set your default hourly rate (config file)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("chronos.json")
		if err != nil {
			return err
		}
		rate, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return err
		}
		cfg.DefaultRate = rate
		return config.SaveConfig("chronos.json", cfg)
	},
}

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "Show client/project analytics (limited functionality post-refactor)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("schema init: %w", err)
		}
		entries, err := chronos.ListEntries(dbStore, nil)
		if err != nil {
			return fmt.Errorf("failed to list entries: %w", err)
		}

		log.Info("Project Totals (Hours per ProjectID):")
		projectTotalsMinutes := chronos.CalculateProjectTotalsByProjectID(entries)
		for projectID, totalMinutes := range projectTotalsMinutes {
			// TODO: Enhance output by fetching project name using chronos.GetProjectByID(dbStore, projectID)
			log.Info(fmt.Sprintf("- ProjectID %d: %.2f hours", projectID, totalMinutes/60.0))
		}
		
		log.Warn("Client and Task specific analytics are not available with the current chronos.Entry structure.")
		log.Warn("To re-enable, chronos.Entry would need direct Client/Task fields or advanced parsing/lookups.")
		return nil
	},
}

var reviewCmd = &cobra.Command{
	Use:   "review [period]",
	Short: "Generate a weekly or monthly review (limited functionality post-refactor)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		period := "week"
		if len(args) > 0 {
			period = args[0]
		}
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("schema init: %w", err)
		}

		var sinceFilter time.Time
		if period == "month" {
			sinceFilter = time.Now().AddDate(0, -1, 0)
		} else { // Default to week
			sinceFilter = time.Now().AddDate(0, 0, -7)
		}
		
		// chronos.ListEntries filter for date range: map[string]interface{}{"start_date": sinceFilter}
		// This gets entries that started *on or after* sinceFilter.
		// We might want to filter entries client-side as well if ListEntries doesn't support precise range.
		allEntries, err := chronos.ListEntries(dbStore, nil) // Fetch all, then filter by date client-side for review period
		if err != nil {
			return fmt.Errorf("failed to list entries for review: %w", err)
		}

		totalMinutesInPeriod := chronos.CalculateReviewPeriodTotals(allEntries, sinceFilter)
		
		log.Info(fmt.Sprintf("%s review: %.2f hours", strings.Title(period), totalMinutesInPeriod/60.0))
		log.Warn("LLM summary and detailed earnings/task breakdown need rework due to struct changes.")
		// TODO: Add LLM summary and export as Markdown
		return nil
	},
}

// templateCmd uses direct DB access for a separate 'templates' table, not related to Entry/Block.
var templateCmd = &cobra.Command{
	Use:   "template [name] [entry]",
	Short: "Save or use an entry template/snippet",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("db store: %w", err)
		}
		// Note: InitSchema() might also call EnsureTemplatesTable if we integrate it there,
		// but explicit call from chronos.SaveTemplate/GetTemplate is safer for direct usage.
		// For this command, InitSchema is generally good practice if other parts of dbStore are used.
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("schema init: %w", err)
		}

		if len(args) == 1 { // Get template
			name := utils.SanitizeString(args[0])
			entryText, err := chronos.GetTemplate(dbStore, name)
			if err != nil {
				// chronos.GetTemplate already provides a descriptive error, including "not found"
				log.Error("Failed to retrieve template", "name", name, "error", err)
				return err // Return the error directly
			}
			log.Info("Template loaded", "name", name, "text", entryText)
			// TODO: Consider what to do with the template text, e.g., pre-fill add command, copy to clipboard.
			// For now, just logging it.
			return nil
		}
		if len(args) >= 2 { // Save template
			name := utils.SanitizeString(args[0])
			entryText := utils.SanitizeDescription(strings.Join(args[1:], " "))

			if err := chronos.SaveTemplate(dbStore, name, entryText); err != nil {
				log.Error("Failed to save template", "name", name, "error", err)
				return err // Return the error directly
			}
			log.Info("Template saved successfully.", "name", name)
			return nil
		}
		return fmt.Errorf("incorrect number of arguments for template command")
	},
}

var invoiceSmartCmd = &cobra.Command{
	Use:   "invoice-smart",
	Short: "Detect unbilled entries and mark as invoiced",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("schema init: %w", err)
		}

		// Find unbilled entries
		unbilledEntries, err := chronos.ListEntries(dbStore, map[string]interface{}{"invoiced": false})
		if err != nil {
			return fmt.Errorf("failed to find unbilled entries: %w", err)
		}

		if len(unbilledEntries) == 0 {
			log.Info("No unbilled entries found.")
			return nil
		}

		var successfullyMarkedCount int
		var errorMessages []string

		for _, entry := range unbilledEntries {
			entry.Invoiced = true
			entry.UpdatedAt = time.Now()
			if errUpdate := chronos.UpdateEntry(dbStore, entry); errUpdate != nil {
				errMsg := fmt.Sprintf("Failed to mark entry ID %d as invoiced: %v", entry.ID, errUpdate)
				log.Error(errMsg)
				errorMessages = append(errorMessages, errMsg)
			} else {
				successfullyMarkedCount++
			}
		}

		log.Info(fmt.Sprintf("Marked %d entries as invoiced.", successfullyMarkedCount))
		if len(errorMessages) > 0 {
			log.Error("Some entries could not be marked as invoiced:")
			for _, msg := range errorMessages {
				log.Error(fmt.Sprintf("- %s", msg))
			}
			return fmt.Errorf("%d entries failed to update", len(errorMessages))
		}
		return nil
	},
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check Chronos dependencies and environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("Chronos Doctor: Checking system dependencies...")
		// Check Ollama
		ollamaPath, err := exec.LookPath("ollama")
		if err != nil {
			log.Error("Ollama not found in PATH.")
			log.Info("Install Ollama: https://ollama.com/download or use your package manager.")
		} else {
			log.Info("Ollama found: %s", ollamaPath)
			// Try running ollama list
			_, err := exec.Command("ollama", "list").CombinedOutput()
			if err != nil {
				log.Error("Ollama is installed but not responding. Try restarting the Ollama service.")
			} else {
				log.Info("Ollama is responding.")
			}
		}
		// Check SQLite3
		_, errSqlite := exec.LookPath("sqlite3")
		if errSqlite != nil {
			log.Error("sqlite3 CLI not found in PATH.")
			log.Info("Install SQLite: sudo apt install sqlite3 (Linux), brew install sqlite3 (macOS), or see https://sqlite.org/download.html")
		} else {
			log.Info("sqlite3 CLI found.")
		}
		// Check DB file
		if _, errDbFile := os.Stat("chronos.db"); os.IsNotExist(errDbFile) {
			log.Warn("chronos.db not found. It will be created on first use.")
		} else {
			log.Info("chronos.db found.")
		}
		// Check .env
		envPath := config.FindEnvPath()
		if _, errEnvFile := os.Stat(envPath); os.IsNotExist(errEnvFile) {
			log.Warn(fmt.Sprintf(".env not found at %s. Run 'chronos' to create one interactively or ensure it's in the correct location.", envPath))
		} else {
			log.Info(fmt.Sprintf(".env config found at %s.", envPath))
		}
		log.Info("Doctor check complete.")
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Set Charm log to Info level for unified output
	log.SetLevel(log.InfoLevel)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle") // Example global flag
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(suggestCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(remindCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(completeCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(pomodoroCmd)
	rootCmd.AddCommand(idleCmd)
	rootCmd.AddCommand(rateCmd)
	rootCmd.AddCommand(analyticsCmd)
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(invoiceSmartCmd)
	rootCmd.AddCommand(doctorCmd)

	// Check for .env and prompt if missing
	// This runs when cmd package is initialized.
	go func() {
		if _, err := os.Stat(config.FindEnvPath()); os.IsNotExist(err) {
			// This runs in a goroutine to avoid blocking if prompt is shown
			// However, cobra execution might proceed before config is ready.
			// Consider a more robust way to handle initial config if critical for all commands.
			// For now, it's a non-blocking check.
			// _ = config.PromptEnvConfig() // This was causing issues with tests, disabling for now.
		}
	}()
}
