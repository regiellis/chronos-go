package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/regiellis/chronos-go/utils"
)

type SummaryViewModel struct {
	Summary string // This could be a struct for richer data
}

func NewSummaryViewModel(summary string) *SummaryViewModel {
	return &SummaryViewModel{Summary: summary}
}

func (m *SummaryViewModel) Init() tea.Cmd {
	return nil
}

func (m *SummaryViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *SummaryViewModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		utils.TitleStyle.Render("Summary"),
		utils.ValueStyle.Render(m.Summary),
		"",
		utils.InactiveStyle.Render("Press q to quit"),
	)
}
