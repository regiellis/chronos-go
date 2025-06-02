package ui

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/regiellis/chronos-go/chronos"
	"github.com/regiellis/chronos-go/utils"
)

type EntryViewModel struct {
	Entry *chronos.Entry
}

func NewEntryViewModel(entry *chronos.Entry) *EntryViewModel {
	return &EntryViewModel{Entry: entry}
}

func (m *EntryViewModel) Init() tea.Cmd {
	return nil
}

func (m *EntryViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *EntryViewModel) View() string {
	if m.Entry == nil {
		return utils.ErrorStyle.Render("No entry selected.")
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		utils.TitleStyle.Render("Entry Details"),
		utils.LabelStyle.Render("Project: ")+utils.ValueStyle.Render(utils.SanitizeString(m.Entry.Project)),
		utils.LabelStyle.Render("Client: ")+utils.ValueStyle.Render(utils.SanitizeString(m.Entry.Client)),
		utils.LabelStyle.Render("Task: ")+utils.ValueStyle.Render(utils.SanitizeString(m.Entry.Task)),
		utils.LabelStyle.Render("Description: ")+utils.ValueStyle.Render(utils.SanitizeDescription(m.Entry.Description)),
		utils.LabelStyle.Render("Duration: ")+utils.ValueStyle.Render(fmt.Sprintf("%d min", m.Entry.Duration)),
		utils.LabelStyle.Render("Date: ")+utils.ValueStyle.Render(m.Entry.EntryTime.Format("2006-01-02 15:04")),
		"",
		utils.InactiveStyle.Render("Press q to quit"),
	)
}

type EntryFormModel struct {
	Form        *huh.Form
	Entry       *chronos.Entry
	Completed   bool
	Suggestion  string
	DurationStr string
	RateStr     string
}

func NewEntryFormModel(suggestion string) *EntryFormModel {
	entry := &chronos.Entry{}
	model := &EntryFormModel{Entry: entry, Suggestion: suggestion}
	model.Form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Project").Value(&entry.Project),
			huh.NewInput().Title("Client").Value(&entry.Client),
			huh.NewInput().Title("Task").Value(&entry.Task),
			huh.NewInput().Title("Description").Value(&entry.Description),
			huh.NewInput().Title("Duration (min)").Value(&model.DurationStr),
			huh.NewConfirm().Title("Billable?").Value(&entry.Billable),
			huh.NewInput().Title("Rate (per hour)").Value(&model.RateStr),
		),
	)
	return model
}

func (m *EntryFormModel) Init() tea.Cmd {
	return m.Form.Init()
}

func (m *EntryFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.Form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.Form = f
	}
	if m.Form.State == huh.StateCompleted {
		m.Completed = true
		// Parse duration and rate
		if d, err := strconv.ParseInt(m.DurationStr, 10, 64); err == nil {
			m.Entry.Duration = d
		}
		if r, err := strconv.ParseFloat(m.RateStr, 64); err == nil {
			m.Entry.Rate = r
		}
		// Sanitize fields before saving
		m.Entry.Project = utils.SanitizeString(m.Entry.Project)
		m.Entry.Client = utils.SanitizeString(m.Entry.Client)
		m.Entry.Task = utils.SanitizeString(m.Entry.Task)
		m.Entry.Description = utils.SanitizeDescription(m.Entry.Description)
	}
	return m, cmd
}

func (m *EntryFormModel) View() string {
	v := lipgloss.JoinVertical(lipgloss.Left,
		utils.TitleStyle.Render("Add Entry (AI Suggestion: "+m.Suggestion+")"),
		m.Form.View(),
	)
	if m.Completed {
		v += utils.ActiveStyle.Render("Entry saved!")
	}
	return v
}
