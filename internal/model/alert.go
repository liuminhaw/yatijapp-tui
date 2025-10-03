package model

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type Alert struct {
	title     string
	alertType string // "confirmation", "notification"

	prompts  []string
	warnings []string
	cmds     map[string]tea.Cmd

	width int
}

func NewAlert(
	title, alertType string,
	prompts, warnings []string,
	width int,
	cmds map[string]tea.Cmd,
) Alert {
	return Alert{
		title:     title,
		alertType: alertType,
		prompts:   prompts,
		warnings:  warnings,
		cmds:      cmds,
		width:     width,
	}
}

func (a Alert) Init() tea.Cmd {
	return nil
}

func (a Alert) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch a.alertType {
	case "confirmation":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				return a, a.cmds["confirm"]
			case "n", "N":
				return a, a.cmds["cancel"]
			}
		}
	case "notification":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "<":
				return a, a.cmds["return"]
			}
		}
	default:
		panic("unknown alert type: " + a.alertType)
	}

	return a, nil
}

func (a Alert) View() string {
	var helper string
	switch a.alertType {
	case "confirmation":
		helper = lipgloss.StyleRanges(
			"[y]es        [n]o",
			lipgloss.Range{Start: 0, End: 3, Style: style.Document.Primary},
			lipgloss.Range{Start: 3, End: 5, Style: style.Document.Normal},
			lipgloss.Range{Start: 13, End: 16, Style: style.Document.Primary},
			lipgloss.Range{Start: 16, End: 17, Style: style.Document.Normal},
		)
	case "notification":
		helper = lipgloss.StyleRanges(
			"< back",
			lipgloss.Range{Start: 0, End: 1, Style: style.Document.Primary},
			lipgloss.Range{Start: 2, End: 6, Style: style.Document.Normal},
		)
	default:
		panic("unknown alert type: " + a.alertType)
	}

	var b strings.Builder
	b.WriteString(style.Document.Secondary.Bold(true).Render(a.title) + "\n\n")
	for _, prompt := range a.prompts {
		b.WriteString(style.Document.Highlight.Render(prompt) + "\n")
	}
	for _, warning := range a.warnings {
		b.WriteString(style.ErrorStyle.Render(warning) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helper)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().
			Width(a.width).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.Text).
			Padding(0, 1).
			Render(b.String()),
	)
}
