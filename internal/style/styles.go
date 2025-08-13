package style

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
)

var DocumentStyle = struct {
	Highlight lipgloss.Style
	Normal    lipgloss.Style
}{
	Highlight: lipgloss.NewStyle().Foreground(colors.DocumentText).Bold(true),
	Normal:    lipgloss.NewStyle().Foreground(colors.DocumentTextDim),
}

var (
	MsgStyle     lipgloss.Style = lipgloss.NewStyle().Foreground(colors.MsgText).Bold(true)
	WarningStyle lipgloss.Style = lipgloss.NewStyle().Foreground(colors.WarningText).Bold(true)
	ErrorStyle   lipgloss.Style = lipgloss.NewStyle().Foreground(colors.ErrorText).Bold(true)
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
	Prompt: lipgloss.NewStyle().Foreground(colors.DocumentText).Bold(true),
	// Selected: lipgloss.NewStyle().Foreground(colors.SelectionText).Bold(true),
	Selected: lipgloss.NewStyle().Foreground(colors.DocumentTextBright).Bold(true),
	Document: lipgloss.NewStyle().Foreground(colors.DocumentText),
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
	Choice        lipgloss.Style
	ChoiceContent lipgloss.Style
}{
	"default": {
		Choices: lipgloss.NewStyle().Foreground(colors.DocumentText),
		Choice:  lipgloss.NewStyle().Foreground(colors.HighlightText).Bold(true),
	},
	"list": {
		Choices: lipgloss.NewStyle().Foreground(colors.DocumentText),
		Choice: lipgloss.NewStyle().
			Foreground(colors.MainText).
			Background(colors.MainBg).
			Bold(true),
		ChoiceContent: lipgloss.NewStyle().Foreground(colors.DocumentText),
	},
}

var BorderStyling lipgloss.Style = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(colors.DocumentTextDim)

var FormFieldStyle = struct {
	Content lipgloss.Style
	Helper  lipgloss.Style
	Error   lipgloss.Style
	// Prompt  func(bool) lipgloss.Style
	Prompt func(string, bool) string
}{
	Content: InputStyle.Document,
	Helper:  InputStyle.Helper,
	Error:   ErrorStyle,
	// Prompt: func(focused bool) lipgloss.Style {
	// 	if focused {
	// 		return InputStyle.Selected
	// 	}
	// 	return InputStyle.Prompt
	// },
	Prompt: func(s string, focused bool) string {
		if focused {
			return InputStyle.Selected.Render(fmt.Sprintf("‣ %s", s))
		}
		return InputStyle.Prompt.Render(s)
	},
}

func ContainerStyle(terminalWidth int, container string, marginHeight int) lipgloss.Style {
	containerWidth := lipgloss.Width(container)
	containerWidthMargin := (terminalWidth - containerWidth) / 2

	return lipgloss.NewStyle().Margin(marginHeight, containerWidthMargin, 0)
}
