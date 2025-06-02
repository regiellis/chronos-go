package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/regiellis/chronos-go/chronos"
	"github.com/regiellis/chronos-go/db"
	"github.com/regiellis/chronos-go/utils"
)

type AppViewModel struct {
	Current tea.Model
}

func NewAppViewModel(initial tea.Model) *AppViewModel {
	return &AppViewModel{Current: initial}
}

func (m *AppViewModel) Init() tea.Cmd {
	return m.Current.Init()
}

func (m *AppViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.Current.Update(msg)
	m.Current = model
	return m, cmd
}

func (m *AppViewModel) View() string {
	return lipgloss.NewStyle().Padding(1, 2).Render(m.Current.View())
}

type PomodoroModel struct {
	Duration  time.Duration
	Remaining time.Duration
	Running   bool
	Completed bool
	StartTime time.Time
}

func NewPomodoroModel(dur time.Duration) *PomodoroModel {
	return &PomodoroModel{
		Duration:  dur,
		Remaining: dur,
		Running:   true,
		StartTime: time.Now(),
	}
}

func (m *PomodoroModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type tickMsg time.Time

func (m *PomodoroModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.Running && !m.Completed {
			m.Remaining -= time.Second
			if m.Remaining <= 0 {
				m.Completed = true
				m.Running = false
				// Log the session as an entry (default project/task)
				go logPomodoroEntry(m.Duration)
				return m, nil
			}
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.Running = false
			return m, tea.Quit
		}
	}
	return m, nil
}

func logPomodoroEntry(dur time.Duration) {
	dbStore, err := db.NewStore("chronos.db")
	if err != nil {
		return
	}
	dbStore.InitSchema()
	entry := &chronos.Entry{
		Project:     utils.SanitizeString("Pomodoro"),
		Task:        utils.SanitizeString("Focus Session"),
		Description: utils.SanitizeDescription("Pomodoro focus session"),
		Duration:    int64(dur.Minutes()),
		EntryTime:   time.Now(),
		CreatedAt:   time.Now(),
		Billable:    false,
		Rate:        0,
	}
	block, _ := dbStore.GetActiveBlock()
	if block != nil {
		entry.BlockID = block.ID
		entry.Client = utils.SanitizeString(block.Client)
		entry.Project = utils.SanitizeString(block.Project)
	}
	dbStore.AddEntry(entry)
}

func (m *PomodoroModel) View() string {
	if m.Completed {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#859900")).Bold(true).Render("Pomodoro complete! Press q to exit.")
	}
	min := int(m.Remaining.Minutes())
	sec := int(m.Remaining.Seconds()) % 60
	return fmt.Sprintf("Focus: %02d:%02d remaining\nPress q to quit.", min, sec)
}
