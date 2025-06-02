package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/regiellis/chronos-go/chronos"
	"github.com/regiellis/chronos-go/utils"
)

type ListViewModel struct {
	Entries []*chronos.Entry
	Cursor  int
}

func NewListViewModel(entries []*chronos.Entry) *ListViewModel {
	return &ListViewModel{Entries: entries}
}

func (m *ListViewModel) Init() tea.Cmd {
	return nil
}

func (m *ListViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Entries)-1 {
				m.Cursor++
			}
		}
	}
	return m, nil
}

func (m *ListViewModel) View() string {
	if len(m.Entries) == 0 {
		return utils.ErrorStyle.Render("No entries found.")
	}
	var rows []string
	for i, e := range m.Entries {
		cursor := "  "
		style := utils.InactiveStyle
		if i == m.Cursor {
			cursor = "> "
			style = utils.ActiveStyle
		}
		row := fmt.Sprintf("%s%s | %s | %s | %d min | %s", cursor, utils.SanitizeString(e.Project), utils.SanitizeString(e.Task), utils.SanitizeDescription(e.Description), e.Duration, e.EntryTime.Format("2006-01-02"))
		rows = append(rows, style.Render(row))
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		utils.TitleStyle.Render("Entries List"),
		lipgloss.JoinVertical(lipgloss.Left, rows...),
		"",
		utils.InactiveStyle.Render("↑/↓ to navigate, q to quit"),
	)
}
