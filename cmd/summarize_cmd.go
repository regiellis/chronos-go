package cmd

import (
	"fmt"

	"github.com/regiellis/chronos-go/db"
	"github.com/regiellis/chronos-go/llm"
	"github.com/regiellis/chronos-go/utils"
	"github.com/spf13/cobra"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "Get an LLM-generated summary for a block",
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
		entries, err := dbStore.ListEntries(map[string]interface{}{"block_id": block.ID})
		if err != nil {
			return err
		}
		llmClient := &llm.OllamaClient{Model: "llama2:7b"}
		summary, err := llmClient.SummarizeBlock(block, entries)
		if err != nil {
			return err
		}
		fmt.Println(utils.TitleStyle.Render("Block Summary"))
		fmt.Println(utils.LLMStyle.Render(summary))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
}
