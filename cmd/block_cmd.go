package cmd

import (
	"fmt"
	"time"

	"github.com/regiellis/chronos-go/chronos" // Imported chronos
	"github.com/regiellis/chronos-go/db"
	"github.com/regiellis/chronos-go/utils"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Manage time blocks (sprints, projects, etc)",
}

var blockStartCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start a new block and mark it active",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("failed to initialize schema: %w", err)
		}

		durStr, _ := cmd.Flags().GetString("duration")
		client, _ := cmd.Flags().GetString("client")
		project, _ := cmd.Flags().GetString("project")

		startTime := time.Now()
		endTime := startTime.Add(14 * 24 * time.Hour) // Default 2 weeks
		if durStr != "" {
			parsedDur, errDur := time.ParseDuration(durStr)
			if errDur == nil {
				endTime = startTime.Add(parsedDur)
			} else {
				fmt.Printf("Warning: could not parse duration '%s', using default: %v\n", durStr, errDur)
			}
		}

		newBlock := &chronos.Block{
			Name:      utils.SanitizeString(args[0]),
			Client:    utils.SanitizeString(client), // Consider mapping to ClientID
			Project:   utils.SanitizeString(project), // Consider mapping to ProjectID
			StartTime: startTime,
			EndTime:   endTime,
			Active:    false, // Will be set to true by SetActiveBlock
			// CreatedAt and UpdatedAt are set by chronos.CreateBlock
		}

		// Create the block using the new chronos function
		if err := chronos.CreateBlock(dbStore, newBlock); err != nil {
			return fmt.Errorf("failed to create block: %w", err)
		}

		// Set this new block as active using its ID from newBlock.ID
		if newBlock.ID == 0 {
			return fmt.Errorf("created block ID is 0, cannot set active")
		}
		if err := chronos.SetActiveBlock(dbStore, newBlock.ID); err != nil {
			// If setting active fails, we might want to inform the user the block was created but not activated.
			// For now, return the error.
			return fmt.Errorf("block created (ID: %d) but failed to set as active: %w", newBlock.ID, err)
		}
		
		// Update newBlock.Active to true as SetActiveBlock was successful
		newBlock.Active = true


		fmt.Println(utils.SuccessStyle.Render("Started new block and set as active!"))
		fmt.Println(utils.EntryStyle.Render(
			fmt.Sprintf("ID: %d\nName: %s\nClient: %s\nProject: %s\nStart: %s\nEnd: %s\nActive: %t",
				newBlock.ID, newBlock.Name, newBlock.Client, newBlock.Project,
				newBlock.StartTime.Format("2006-01-02"), newBlock.EndTime.Format("2006-01-02"), newBlock.Active,
			),
		))
		return nil
	},
}

// TODO: Add commands for block stop, list, etc., using chronos package functions.
// Example: blockStopCmd
var blockStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the currently active block",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := "chronos.db"
		dbStore, err := db.NewStore(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create db store: %w", err)
		}
		if err := dbStore.InitSchema(); err != nil {
			return fmt.Errorf("failed to initialize schema: %w", err)
		}

		activeBlock, err := chronos.GetActiveBlock(dbStore)
		if err != nil {
			return fmt.Errorf("failed to get active block: %w", err)
		}
		if activeBlock == nil {
			fmt.Println(utils.InfoStyle.Render("No block is currently active."))
			return nil
		}

		activeBlock.Active = false
		activeBlock.EndTime = time.Now() // Optionally set EndTime to now when stopping
		if err := chronos.UpdateBlock(dbStore, activeBlock); err != nil {
			return fmt.Errorf("failed to stop block ID %d: %w", activeBlock.ID, err)
		}

		fmt.Println(utils.SuccessStyle.Render(fmt.Sprintf("Stopped block: %s (ID: %d)", activeBlock.Name, activeBlock.ID)))
		return nil
	},
}


func init() {
	blockStartCmd.Flags().String("duration", "2w", "Block duration (e.g. 2w, 10d, 1m)")
	blockStartCmd.Flags().String("client", "", "Client name (optional)")
	blockStartCmd.Flags().String("project", "", "Project name (optional)")
	blockCmd.AddCommand(blockStartCmd)
	blockCmd.AddCommand(blockStopCmd) // Add new stop command
	rootCmd.AddCommand(blockCmd)
}
