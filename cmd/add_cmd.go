package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/regiellis/chronos-go/chronos"
	"github.com/regiellis/chronos-go/db"
	"github.com/regiellis/chronos-go/llm"
	"github.com/regiellis/chronos-go/utils"
	"github.com/spf13/cobra"
)

var (
	addScale      string
	addScaleCount int
	addScaleLeft  int
	addSuggest    bool
	addLLM        bool
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [entry]",
	Short: "Add a time entry (optionally with --scale)",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return err
		}
		if err := dbStore.InitSchema(); err != nil {
			return err
		}
		llmClient := &llm.OllamaClient{Model: "llama2:7b"}
		if addSuggest || len(args) == 0 {
			entries, _ := dbStore.ListEntries(nil)
			blocks, _ := dbStore.ListBlocks(nil)
			if addSuggest {
				suggestion, err := llmClient.SuggestNextEntry(entries, blocks)
				if err == nil && suggestion != "" {
					fmt.Println(utils.LLMStyle.Render(suggestion))
				}
			}
			if len(args) == 0 {
				return nil
			}
		}
		input := args[0]
		// Only use LLM to parse entry if --llm flag is set, otherwise require structured/manual entry parsing
		useLLM, _ := cmd.Flags().GetBool("llm")
		var entry *chronos.Entry
		if useLLM {
			entry, err = llmClient.ParseEntry(input)
			if err != nil {
				return err
			}
		} else {
			// Manual fallback: parse input as "duration project task description"
			// Accepts duration in minutes (int) or as Go duration string (e.g. 1h, 30m)
			var duration int64
			var project, task, description string
			parts := make([]string, 0)
			for _, part := range strings.SplitN(input, " ", 4) {
				if part != "" {
					parts = append(parts, part)
				}
			}
			if len(parts) < 4 {
				return fmt.Errorf("Could not parse entry. Use --llm or provide: <duration> <project> <task> <description>")
			}
			// Try parsing duration as int (minutes), then as Go duration string
			// Accept both uppercase and lowercase units (e.g., 1H, 30M)
			durStr := strings.ToLower(parts[0])
			duration, err = strconv.ParseInt(durStr, 10, 64)
			if err != nil {
				parsed, err2 := time.ParseDuration(durStr)
				if err2 != nil {
					return fmt.Errorf("Invalid duration: %v (tried int and Go duration string)", err2)
				}
				duration = int64(parsed.Minutes())
			}
			project = parts[1]
			task = parts[2]
			description = parts[3]
			entry = &chronos.Entry{
				Project:     project,
				Task:        task,
				Description: description,
				Duration:    duration,
				EntryTime:   time.Now(),
				CreatedAt:   time.Now(),
				Billable:    true,
				Rate:        0,
			}
		}
		// Sanitize all user/LLM fields
		entry.Project = utils.SanitizeString(entry.Project)
		entry.Client = utils.SanitizeString(entry.Client)
		entry.Task = utils.SanitizeString(entry.Task)
		entry.Description = utils.SanitizeDescription(entry.Description)
		// Handle --scale and --scale-next
		if addScale != "" && (addScaleCount == 0 || addScaleLeft > 0) {
			parsed, err := time.ParseDuration(addScale)
			if err == nil {
				entry.Duration = int64(parsed.Minutes())
			}
			if addScaleCount > 0 {
				addScaleLeft--
			}
		}
		// Set entry time to now if not provided
		if entry.EntryTime.IsZero() {
			entry.EntryTime = time.Now()
		}
		entry.CreatedAt = time.Now()
		// Attach to active block if present
		block, _ := dbStore.GetActiveBlock()
		if block != nil {
			entry.BlockID = block.ID
			if entry.Client == "" {
				entry.Client = block.Client
			}
			if entry.Project == "" {
				entry.Project = block.Project
			}
		}
		fmt.Println("DEBUG: About to call AddEntry with:", entry)
		if err := dbStore.AddEntry(entry); err != nil {
			fmt.Println("DEBUG: AddEntry failed:", err)
			return err
		}
		// Themed user feedback: show the entry just added
		fmt.Println(utils.SuccessStyle.Render("Entry added!"))
		fmt.Println(utils.EntryStyle.Render(fmt.Sprintf(
			"Project: %s\nClient: %s\nTask: %s\nDescription: %s\nDuration: %d min\nTime: %s",
			entry.Project, entry.Client, entry.Task, entry.Description, entry.Duration, entry.EntryTime.Format("2006-01-02 15:04"))))
		// Only call LLM if --llm flag is set
		useLLM, _ = cmd.Flags().GetBool("llm")
		if useLLM {
			llmClient := llm.NewOllamaClient()
			entries, _ := dbStore.ListEntries(nil)
			feedback, err := llmClient.FeedbackAfterEntry(entry, entries)
			if err == nil && feedback != "" {
				fmt.Println(utils.LLMStyle.Render(feedback))
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVar(&addScale, "scale", "", "Override duration for this entry (e.g. 1h, 30m, 15m)")
	addCmd.Flags().IntVar(&addScaleCount, "scale-next", 0, "Apply scale to the next N entries")
	addCmd.Flags().BoolVar(&addSuggest, "suggest", false, "Show LLM-powered suggestions before entry")
	addCmd.Flags().BoolVar(&addLLM, "llm", false, "Use LLM for feedback after entry")
}
