package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/regiellis/chronos-go/db"
	"github.com/regiellis/chronos-go/ui"
)

func main() {
	dbStore, err := db.NewStore("chronos.db")
	if err != nil {
		panic(err)
	}
	if err := dbStore.InitSchema(); err != nil {
		panic(err)
	}
	model := ui.NewMainMenuModel(dbStore)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
