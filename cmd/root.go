/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"
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
			return err
		}
		if err := dbStore.InitSchema(); err != nil {
			return err
		}
		llmClient := llm.NewOllamaClient()
		question := strings.Join(args, " ")
		entries, _ := dbStore.ListEntries(nil)
		blocks, _ := dbStore.ListBlocks(nil)
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
			return err
		}
		if err := dbStore.InitSchema(); err != nil {
			return err
		}
		llmClient := llm.NewOllamaClient()
		entries, _ := dbStore.ListEntries(nil)
		blocks, _ := dbStore.ListBlocks(nil)
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
		return askCmd.RunE(cmd, args)
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
			return err
		}
		if err := dbStore.InitSchema(); err != nil {
			return err
		}
		llmClient := llm.NewOllamaClient()
		entries, _ := dbStore.ListEntries(nil)
		blocks, _ := dbStore.ListBlocks(nil)
		reminder, err := llmClient.SmartReminder(entries, blocks)
		if err != nil {
			return err
		}
		log.Info(reminder)
		return nil
	},
}

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show your recent LLM queries for quick re-use",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
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
			return err
		}
		llmClient := llm.NewOllamaClient()
		entries, _ := dbStore.ListEntries(nil)
		blocks, _ := dbStore.ListBlocks(nil)
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
	Short: "Edit a time entry by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return err
		}
		entries, err := dbStore.ListEntries(map[string]interface{}{"id": id})
		if err != nil || len(entries) == 0 {
			log.Error("Entry not found")
			return fmt.Errorf("Entry not found")
		}
		entry := entries[0]
		// For demo: just toggle billable
		entry.Billable = !entry.Billable
		return dbStore.UpdateEntry(entry)
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
			return err
		}
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return err
		}
		return dbStore.DeleteEntry(id)
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
		model := ui.NewPomodoroModel(dur)
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
			return err
		}
		entries, err := dbStore.ListEntries(nil)
		if err != nil {
			return err
		}
		if len(entries) < 2 {
			log.Warn("Not enough entries to detect idle gaps.")
			return nil
		}
		// Sort by EntryTime ascending
		for i := 1; i < len(entries); i++ {
			gap := entries[i].EntryTime.Sub(entries[i-1].EntryTime)
			if gap > 2*time.Hour {
				log.Warn("Idle gap detected: %s to %s (%v)", entries[i-1].EntryTime.Format("2006-01-02 15:04"), entries[i].EntryTime.Format("2006-01-02 15:04"), gap)
			}
		}
		return nil
	},
}

var rateCmd = &cobra.Command{
	Use:   "rate [amount]",
	Short: "Quickly set your default hourly rate",
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
	Short: "Show client/project analytics (top clients, most frequent tasks, etc)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		entries, err := dbStore.ListEntries(nil)
		if err != nil {
			return err
		}
		clientTotals := map[string]int64{}
		projectTotals := map[string]int64{}
		taskTotals := map[string]int64{}
		for _, e := range entries {
			clientTotals[e.Client] += e.Duration
			projectTotals[e.Project] += e.Duration
			taskTotals[e.Task] += e.Duration
		}
		log.Info("Top Clients:")
		for c, d := range clientTotals {
			log.Info("- %s: %.2f hours", c, float64(d)/60.0)
		}
		log.Info("")
		log.Info("Top Projects:")
		for p, d := range projectTotals {
			log.Info("- %s: %.2f hours", p, float64(d)/60.0)
		}
		log.Info("")
		log.Info("Top Tasks:")
		for t, d := range taskTotals {
			log.Info("- %s: %.2f hours", t, float64(d)/60.0)
		}
		return nil
	},
}

var reviewCmd = &cobra.Command{
	Use:   "review [period]",
	Short: "Generate a weekly or monthly review (hours, earnings, top tasks, LLM summary)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		period := "week"
		if len(args) > 0 {
			period = args[0]
		}
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		entries, err := dbStore.ListEntries(nil)
		if err != nil {
			return err
		}
		var start time.Time
		if period == "month" {
			start = time.Now().AddDate(0, -1, 0)
		} else {
			start = time.Now().AddDate(0, 0, -7)
		}
		var totalMinutes int64
		for _, e := range entries {
			if e.EntryTime.After(start) {
				totalMinutes += e.Duration
			}
		}
		log.Info("%s review: %.2f hours", period, float64(totalMinutes)/60.0)
		// TODO: Add LLM summary and export as Markdown
		return nil
	},
}

var templateCmd = &cobra.Command{
	Use:   "template [name] [entry]",
	Short: "Save or use an entry template/snippet",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		if len(args) == 1 {
			// Use template
			name := utils.SanitizeString(args[0])
			row := dbStore.DB.QueryRow(`SELECT entry FROM templates WHERE name=?`, name)
			var entryText string
			if err := row.Scan(&entryText); err != nil {
				log.Error("Template not found")
				return fmt.Errorf("Template not found")
			}
			log.Info("Template: %s", entryText)
			return nil
		}
		if len(args) >= 2 {
			// Save template
			name := utils.SanitizeString(args[0])
			entryText := utils.SanitizeDescription(args[1])
			_, err := dbStore.DB.Exec(`CREATE TABLE IF NOT EXISTS templates (name TEXT PRIMARY KEY, entry TEXT)`)
			if err != nil {
				return err
			}
			_, err = dbStore.DB.Exec(`INSERT OR REPLACE INTO templates (name, entry) VALUES (?, ?)`, name, entryText)
			if err != nil {
				return err
			}
			log.Info("Template saved.")
			return nil
		}
		return nil
	},
}

var invoiceSmartCmd = &cobra.Command{
	Use:   "invoice-smart",
	Short: "Detect unbilled entries and mark as invoiced",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		entries, err := dbStore.FindUnbilledEntries()
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			log.Info("No unbilled entries found.")
			return nil
		}
		var ids []int64
		for _, e := range entries {
			ids = append(ids, e.ID)
		}
		err = dbStore.MarkEntriesInvoiced(ids)
		if err != nil {
			return err
		}
		log.Info("Marked %d entries as invoiced.", len(ids))
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
		_, err = exec.LookPath("sqlite3")
		if err != nil {
			log.Error("sqlite3 CLI not found in PATH.")
			log.Info("Install SQLite: sudo apt install sqlite3 (Linux), brew install sqlite3 (macOS), or see https://sqlite.org/download.html")
		} else {
			log.Info("sqlite3 CLI found.")
		}
		// Check DB file
		if _, err := os.Stat("chronos.db"); os.IsNotExist(err) {
			log.Warn("chronos.db not found. It will be created on first use.")
		} else {
			log.Info("chronos.db found.")
		}
		// Check .env
		if _, err := os.Stat(config.FindEnvPath()); os.IsNotExist(err) {
			log.Warn(".env not found. Run 'chronos' to create one interactively.")
		} else {
			log.Info(".env config found.")
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chronos-go.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
	if _, err := os.Stat(config.FindEnvPath()); os.IsNotExist(err) {
		_ = config.PromptEnvConfig()
	}
}
