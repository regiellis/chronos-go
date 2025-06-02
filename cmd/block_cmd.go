package cmd

import (
	"fmt"
	"time"

	"github.com/regiellis/chronos-go/chronos"
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
	Short: "Start a new block",
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
		dur, _ := cmd.Flags().GetString("duration")
		client, _ := cmd.Flags().GetString("client")
		project, _ := cmd.Flags().GetString("project")
		start := time.Now()
		end := start.Add(14 * 24 * time.Hour) // default 2w
		if dur != "" {
			parsed, err := time.ParseDuration(dur)
			if err == nil {
				end = start.Add(parsed)
			}
		}
		block := &chronos.Block{
			Name:      utils.SanitizeString(args[0]),
			Client:    utils.SanitizeString(client),
			Project:   utils.SanitizeString(project),
			StartTime: start,
			EndTime:   end,
			Active:    true,
			CreatedAt: start,
		}
		if err := dbStore.AddBlock(block); err != nil {
			return err
		}
		// Set this block as active
		row := dbStore.DB.QueryRow("SELECT last_insert_rowid()")
		var blockID int64
		if err := row.Scan(&blockID); err == nil {
			dbStore.SetActiveBlock(blockID)
		}
		fmt.Println(utils.SuccessStyle.Render("Started new block!"))
		fmt.Println(utils.EntryStyle.Render(
			fmt.Sprintf("Name: %s\nClient: %s\nProject: %s\nStart: %s\nEnd: %s",
				block.Name, block.Client, block.Project,
				block.StartTime.Format("2006-01-02"), block.EndTime.Format("2006-01-02"),
			),
		))
		return nil
	},
}

func init() {
	blockStartCmd.Flags().String("duration", "2w", "Block duration (e.g. 2w, 10d, 1m)")
	blockStartCmd.Flags().String("client", "", "Client name")
	blockStartCmd.Flags().String("project", "", "Project name")
	blockCmd.AddCommand(blockStartCmd)
	rootCmd.AddCommand(blockCmd)
}
