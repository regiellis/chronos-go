package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/regiellis/chronos-go/db"
)

type MainMenuModel struct {
	Cursor  int
	Choices []string
	DB      *db.Store
}

func NewMainMenuModel(dbStore *db.Store) *MainMenuModel {
	return &MainMenuModel{
		Choices: []string{"View Entries", "Add Entry", "View Blocks", "Summaries", "Quit"},
		DB:      dbStore,
	}
}

func (m *MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}
		case "enter", " ":
			switch m.Cursor {
			case 0:
				entries, _ := m.DB.ListEntries(nil)
				return NewListViewModel(entries), nil
			case 1:
				// Use the EntryFormModel from entry_view.go, pass empty suggestion for now
				return NewEntryFormModel(""), nil
			case 2:
				blocks, _ := m.DB.ListBlocks(nil)
				if len(blocks) == 0 {
					return NewSummaryViewModel("No blocks found."), nil
				}
				// Show the first block for now (could be a list in the future)
				return NewBlockViewModel(blocks[0]), nil
			case 3:
				return NewSummaryViewModel("Summary coming soon!"), nil
			case 4:
				return m, tea.Quit
			}
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *MainMenuModel) View() string {
	s := lipgloss.NewStyle().Bold(true).Render("Chronos TUI - Main Menu\n\n")
	for i, choice := range m.Choices {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}
		s += cursor + choice + "\n"
	}
	s += "\nUse ↑/↓ or j/k to move, Enter to select, q to quit."
	return s
}
