package cmd

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/regiellis/chronos-go/chronos" // Imported chronos
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
	Use:   "add [entry text or parts]",
	Short: "Add a time entry (optionally with --scale and --llm)",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("failed to initialize schema: %w", err)
		}

		llmClient := llm.NewOllamaClient() // Default model, consider making configurable

		if addSuggest || len(args) == 0 {
			entries, listErr := chronos.ListEntries(dbStore, nil)
			if listErr != nil {
				fmt.Println("Warning: Could not list entries for suggestions:", listErr)
			}
			blocks, blockErr := chronos.ListBlocks(dbStore, nil)
			if blockErr != nil {
				fmt.Println("Warning: Could not list blocks for suggestions:", blockErr)
			}

			if addSuggest && entries != nil && blocks != nil { // Ensure we have data for suggestions
				suggestion, llmErr := llmClient.SuggestNextEntry(entries, blocks)
				if llmErr == nil && suggestion != "" {
					fmt.Println(utils.LLMStyle.Render(suggestion))
				} else if llmErr != nil {
					fmt.Println("LLM Suggestion Error:", llmErr)
				}
			}
			if len(args) == 0 {
				return nil // No input provided, and suggestions (if any) are shown.
			}
		}

		input := strings.Join(args, " ") // Join all args to form the input string
		var newEntry chronos.Entry

		useLLM, _ := cmd.Flags().GetBool("llm")
		if useLLM {
			// Assuming llmClient.ParseEntry returns a temporary struct that needs mapping
			// This part is highly dependent on the actual structure returned by llm.ParseEntry
			parsedLLMEntry, llmErr := llmClient.ParseEntry(input)
			if llmErr != nil {
				return fmt.Errorf("LLM parsing failed: %w", llmErr)
			}

			newEntry.Summary = utils.SanitizeDescription(parsedLLMEntry.Summary) // Or Description
			newEntry.StartTime = parsedLLMEntry.StartTime
			if newEntry.StartTime.IsZero() { // Default to now if LLM doesn't provide it
				newEntry.StartTime = time.Now()
			}
			// Duration might be provided directly or as EndTime by LLM
			if parsedLLMEntry.EndTime.IsZero() && parsedLLMEntry.Duration > 0 {
				newEntry.EndTime = newEntry.StartTime.Add(time.Duration(parsedLLMEntry.Duration) * time.Minute)
			} else if !parsedLLMEntry.EndTime.IsZero() {
				newEntry.EndTime = parsedLLMEntry.EndTime
			} else {
				// Default duration if LLM provides neither EndTime nor Duration (e.g., 30 mins)
				newEntry.EndTime = newEntry.StartTime.Add(30 * time.Minute)
			}
			
			// TODO: Implement ProjectID lookup based on parsedLLMEntry.ProjectName
			// For now, placeholder. This would ideally involve chronos.GetProjectByName(dbStore, parsedLLMEntry.ProjectName)
			newEntry.ProjectID = 0 // Placeholder
			// Client handling from LLM is also TBD.
		} else {
			// Manual fallback: parse input as "duration project task description"
			parts := make([]string, 0)
			// Use strings.Fields to handle multiple spaces better than strings.SplitN
			splitArgs := strings.Fields(input) 
			if len(splitArgs) < 4 {
				return fmt.Errorf("could not parse entry. Expected format: <duration> <project> <task> <description>. Got: '%s'", input)
			}
			
			var durationMinutes int64
			durStr := strings.ToLower(splitArgs[0])
			parsedIntDur, errConv := strconv.ParseInt(durStr, 10, 64)
			if errConv != nil {
				parsedGoDur, errDur := time.ParseDuration(durStr)
				if errDur != nil {
					return fmt.Errorf("invalid duration '%s': %v (tried int minutes and Go duration string)", durStr, errDur)
				}
				durationMinutes = int64(parsedGoDur.Minutes())
			} else {
				durationMinutes = parsedIntDur
			}

			projectArg := utils.SanitizeString(splitArgs[1])
			taskArg := utils.SanitizeString(splitArgs[2])
			descriptionArg := utils.SanitizeDescription(strings.Join(splitArgs[3:], " ")) // Remainder is description

			newEntry.Summary = fmt.Sprintf("%s: %s", taskArg, descriptionArg)
			newEntry.StartTime = time.Now()

			// Handle --scale for duration override
			if addScale != "" { // Simpler logic: if --scale is set, it overrides parsed duration
				parsedScaleDur, errScale := time.ParseDuration(addScale)
				if errScale == nil {
					durationMinutes = int64(parsedScaleDur.Minutes())
				} else {
					fmt.Printf("Warning: could not parse --scale duration '%s', using parsed duration: %v\n", addScale, errScale)
				}
			}
			// Note: --scale-next logic is removed for simplicity in this refactoring pass,
			// as it adds statefulness that complicates direct CreateEntry calls.
			// It could be reintroduced by managing addScaleLeft at a higher level or within the command loop.

			newEntry.EndTime = newEntry.StartTime.Add(time.Duration(durationMinutes) * time.Minute)
			
			// TODO: Implement ProjectID lookup based on projectArg
			// For now, placeholder: chronos.GetProjectByName(dbStore, projectArg)
			newEntry.ProjectID = 0 // Placeholder
		}

		newEntry.CreatedAt = time.Now()
		newEntry.UpdatedAt = time.Now()
		newEntry.Invoiced = false // Default for new entries

		activeBlock, errBlock := chronos.GetActiveBlock(dbStore)
		if errBlock != nil && errBlock != sql.ErrNoRows { // sql.ErrNoRows is okay, means no active block
			return fmt.Errorf("failed to get active block: %w", errBlock)
		}
		if activeBlock != nil {
			newEntry.BlockID = activeBlock.ID
			// If ProjectID is still 0, and block has a project, try to use it (needs name to ID lookup)
			if newEntry.ProjectID == 0 && activeBlock.Project != "" {
				// TODO: newEntry.ProjectID = chronos.GetProjectByName(dbStore, activeBlock.Project).ID
			}
		}

		if err := chronos.CreateEntry(dbStore, &newEntry); err != nil {
			return fmt.Errorf("failed to create entry using chronos.CreateEntry: %w", err)
		}

		fmt.Println(utils.SuccessStyle.Render("Entry added!"))
		durationOutput := newEntry.EndTime.Sub(newEntry.StartTime).Minutes()
		fmt.Println(utils.EntryStyle.Render(fmt.Sprintf(
			"Summary: %s\nProjectID: %d\nBlockID: %d\nDuration: %.0f min\nTime: %s",
			newEntry.Summary, newEntry.ProjectID, newEntry.BlockID, durationOutput, newEntry.StartTime.Format("2006-01-02 15:04"))))

		if useLLM { // Check flag again, as it might only be for post-processing
			// Ensure llmClient is the same instance or re-initialize if needed
			entries, listErr := chronos.ListEntries(dbStore, nil)
			if listErr != nil {
				fmt.Println("Warning: Could not list entries for LLM feedback:", listErr)
			} else {
				feedback, llmErr := llmClient.FeedbackAfterEntry(&newEntry, entries)
				if llmErr == nil && feedback != "" {
					fmt.Println(utils.LLMStyle.Render(feedback))
				} else if llmErr != nil {
					fmt.Println("LLM Feedback Error:", llmErr)
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVar(&addScale, "scale", "", "Override duration for this entry (e.g. 1h, 30m, 15m)")
	// --scale-next related flags are kept for now but their logic is simplified/partially removed in RunE
	addCmd.Flags().IntVar(&addScaleCount, "scale-next", 0, "Apply scale to the next N entries (functionality limited in refactor)") 
	addCmd.Flags().BoolVar(&addSuggest, "suggest", false, "Show LLM-powered suggestions before entry")
	addCmd.Flags().BoolVar(&addLLM, "llm", false, "Use LLM for parsing entry and/or feedback after entry")
}
