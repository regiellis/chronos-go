package utils

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Solarized color palette
var (
	Base03  = lipgloss.Color("#002b36")
	Base02  = lipgloss.Color("#073642")
	Base01  = lipgloss.Color("#586e75")
	Base00  = lipgloss.Color("#657b83")
	Base0   = lipgloss.Color("#839496")
	Base1   = lipgloss.Color("#93a1a1")
	Base2   = lipgloss.Color("#eee8d5")
	Base3   = lipgloss.Color("#fdf6e3")
	Yellow  = lipgloss.Color("#b58900")
	Orange  = lipgloss.Color("#cb4b16")
	Red     = lipgloss.Color("#dc322f")
	Magenta = lipgloss.Color("#d33682")
	Violet  = lipgloss.Color("#6c71c4")
	Blue    = lipgloss.Color("#268bd2")
	Cyan    = lipgloss.Color("#2aa198")
	Green   = lipgloss.Color("#859900")
)

// Solarized theme styles
var (
	TitleStyle    = lipgloss.NewStyle().Foreground(Yellow).Background(Base03).Bold(true).Padding(0, 1)
	SubtitleStyle = lipgloss.NewStyle().Foreground(Cyan).Background(Base03).Bold(true)
	LabelStyle    = lipgloss.NewStyle().Foreground(Base1).Background(Base02).Bold(true)
	ValueStyle    = lipgloss.NewStyle().Foreground(Base0).Background(Base02)
	BorderStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(Base01)
	ActiveStyle   = lipgloss.NewStyle().Foreground(Green).Bold(true)
	InactiveStyle = lipgloss.NewStyle().Foreground(Base01)
	ErrorStyle    = lipgloss.NewStyle().Foreground(Red).Bold(true)
)

// Themed user feedback styles
var (
	SuccessStyle = lipgloss.NewStyle().Foreground(Green).Bold(true)
	EntryStyle   = lipgloss.NewStyle().Foreground(Base0).Background(Base02).Padding(0, 1)
	LLMStyle     = lipgloss.NewStyle().Foreground(Magenta).Italic(true)
)

// SanitizeString removes unwanted characters from a string, allowing only safe ones for Project, Client, Task fields.
func SanitizeString(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9 _\-]`)
	return re.ReplaceAllString(s, "")
}

// SanitizeDescription allows more characters but strips control chars and dangerous code.
func SanitizeDescription(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(s, "")
	return s
}

// ValidateField checks if a string is non-empty and not just whitespace.
func ValidateField(s string) bool {
	return strings.TrimSpace(s) != ""
}
