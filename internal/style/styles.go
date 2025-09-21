package style

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
)

var Document = struct {
	Primary   lipgloss.Style
	Secondary lipgloss.Style
	Highlight lipgloss.Style
	Normal    lipgloss.Style
	NormalDim lipgloss.Style
}{
	Primary:   lipgloss.NewStyle().Foreground(colors.Primary),
	Secondary: lipgloss.NewStyle().Foreground(colors.Secondary),
	Highlight: lipgloss.NewStyle().Foreground(colors.Text).Bold(true),
	Normal:    lipgloss.NewStyle().Foreground(colors.Text),
	NormalDim: lipgloss.NewStyle().Foreground(colors.TextMuted),
}

var (
	MsgStyle     lipgloss.Style = lipgloss.NewStyle().Foreground(colors.Success).Bold(true)
	WarningStyle lipgloss.Style = lipgloss.NewStyle().Foreground(colors.Warning).Bold(true)
	ErrorStyle   lipgloss.Style = lipgloss.NewStyle().Foreground(colors.Danger).Bold(true)
)

var InputStyle = struct {
	Title    lipgloss.Style
	Prompt   lipgloss.Style
	Selected lipgloss.Style
	Document lipgloss.Style
	Helper   lipgloss.Style
}{
	Title: lipgloss.NewStyle().
		Foreground(colors.MainText).
		Background(colors.MainBg).
		Padding(0, 2).
		Bold(true),
	Prompt:   lipgloss.NewStyle().Foreground(colors.Primary).Bold(true),
	Selected: lipgloss.NewStyle().Foreground(colors.Secondary).Bold(true),
	Document: lipgloss.NewStyle().Foreground(colors.Text),
	Helper:   lipgloss.NewStyle().Foreground(colors.HelperText).Italic(true),
}

var HelperStyle = map[ViewMode]struct {
	Key    lipgloss.Style
	Action lipgloss.Style
}{
	NormalView: {
		Key:    lipgloss.NewStyle().Foreground(colors.HelperText).Italic(true).Bold(true),
		Action: lipgloss.NewStyle().Foreground(colors.HelperTextDim).Italic(true),
	},
	HighlightView: {
		Key: lipgloss.NewStyle().
			Foreground(colors.MainText).
			Bold(true),
		Action: lipgloss.NewStyle().
			Foreground(colors.DocumentTextBright).
			Italic(true),
	},
}

var ChoicesStyle = map[string]struct {
	Choices       lipgloss.Style
	ChoicesDim    lipgloss.Style
	Choice        lipgloss.Style
	ChoiceContent lipgloss.Style
}{
	"default": {
		Choices: lipgloss.NewStyle().Foreground(colors.TextMuted),
		Choice:  lipgloss.NewStyle().Foreground(colors.Text).Bold(true),
	},
	"list": {
		Choices:    lipgloss.NewStyle().Foreground(colors.Text),
		ChoicesDim: lipgloss.NewStyle().Foreground(colors.DocumentTextDim),
		Choice: lipgloss.NewStyle().
			Foreground(colors.BgLight).
			Background(colors.Text).
			Bold(true),
		ChoiceContent: lipgloss.NewStyle().Foreground(colors.DocumentText),
	},
}

var BorderStyling lipgloss.Style = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colors.Border)

var FormFieldStyle = struct {
	Content lipgloss.Style
	Helper  lipgloss.Style
	Error   lipgloss.Style
	Prompt  func(string, bool) string
}{
	Content: InputStyle.Document,
	Helper:  InputStyle.Helper,
	Error:   ErrorStyle,
	Prompt: func(s string, focused bool) string {
		if focused {
			return InputStyle.Selected.Render(fmt.Sprintf("%s", s))
		}
		return InputStyle.Prompt.Render(s)
	},
}

func ContainerStyle(terminalWidth int, container string, marginHeight int) lipgloss.Style {
	containerWidth := lipgloss.Width(container)
	containerWidthMargin := (terminalWidth - containerWidth) / 2

	return lipgloss.NewStyle().Margin(marginHeight, containerWidthMargin, 0)
}

func StatusColor(status string) lipgloss.AdaptiveColor {
	switch status {
	case "queued":
		return colors.Warning
	case "in progress":
		return colors.Info
	case "completed":
		return colors.Success
	case "canceled":
		return colors.Danger
	default:
		return colors.Text
	}
}

func StatusTextStyle(status string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(StatusColor(status))
}

func CalculateGap(width int, fields ...string) (int, int, bool) {
	fieldsLength := 0
	for _, field := range fields {
		fieldsLength += lipgloss.Width(field)
	}

	if fieldsLength >= width {
		return 0, 0, false
	}

	return (width - fieldsLength) / (len(fields) - 1),
		(width - fieldsLength) % (len(fields) - 1), true
}
