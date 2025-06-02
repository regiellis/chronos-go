package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/regiellis/chronos-go/chronos"
	"github.com/regiellis/chronos-go/utils"
)

type BlockViewModel struct {
	Block *chronos.Block
}

func NewBlockViewModel(block *chronos.Block) *BlockViewModel {
	return &BlockViewModel{Block: block}
}

func (m *BlockViewModel) Init() tea.Cmd {
	return nil
}

func (m *BlockViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *BlockViewModel) View() string {
	if m.Block == nil {
		return utils.ErrorStyle.Render("No active block.")
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		utils.TitleStyle.Render("Active Block"),
		utils.LabelStyle.Render("Name: ")+utils.ValueStyle.Render(utils.SanitizeString(m.Block.Name)),
		utils.LabelStyle.Render("Client: ")+utils.ValueStyle.Render(utils.SanitizeString(m.Block.Client)),
		utils.LabelStyle.Render("Project: ")+utils.ValueStyle.Render(utils.SanitizeString(m.Block.Project)),
		utils.LabelStyle.Render("Start: ")+utils.ValueStyle.Render(m.Block.StartTime.Format("2006-01-02")),
		utils.LabelStyle.Render("End: ")+utils.ValueStyle.Render(m.Block.EndTime.Format("2006-01-02")),
		"",
		utils.InactiveStyle.Render("Press q to quit"),
	)
}
