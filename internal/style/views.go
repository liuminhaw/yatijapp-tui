package style

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

type ViewMode int

const (
	NormalView ViewMode = iota
	HighlightView
)

type ViewSize struct {
	Width  int
	Height int
}

// func TitleBarView(topic string, width int) string {
func TitleBarView(content string, width int, msg bool) string {
	var title string
	if msg {
		title = DocumentStyle.Highlight.Render(
			"Yatijapp",
		) + DocumentStyle.Normal.Render(
			" - ",
		) + MsgStyle.Render(
			content,
		)
	} else {
		title = DocumentStyle.Highlight.Render("Yatijapp") + DocumentStyle.Normal.Render(" - "+content)
	}

	return BorderStyling.Width(width).Padding(0, 1).Render(title)
}

func LoadingView(s *spinner.Model, title string, sizing ViewSize) string {
	titleBar := TitleBarView(title, sizing.Width, false)
	helper := HelperView([]HelperContent{{Key: "<", Action: "back"}}, sizing.Width, NormalView)

	msg := DocumentStyle.Normal.Bold(true).Render("loading...")
	s.Spinner = spinner.Line
	s.Style = DocumentStyle.Highlight

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleBar,
		lipgloss.NewStyle().
			Width(sizing.Width).
			Height(sizing.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("%s %s", s.View(), msg)),
		helper,
	)
}

type HelperContent struct {
	Key    string
	Action string
}

func HelperView(contents []HelperContent, width int, mode ViewMode) string {
	var b strings.Builder
	for i, content := range contents {
		b.WriteString(
			fmt.Sprintf(
				"%s %s",
				HelperStyle[mode].Key.Render(content.Key),
				HelperStyle[mode].Action.Render(content.Action),
			),
		)
		if i < len(contents)-1 {
			b.WriteString("\t")
		}
	}
	// b.WriteString("\n")

	return lipgloss.NewStyle().
		Width(width).
		Margin(1, 0).
		AlignHorizontal(lipgloss.Center).
		Render(b.String())
}

type SideBarContent struct {
	Title        string
	Items        map[string]string
	Order        []string // Order of items to maintain display order
	KeyHighlight bool
}

func SideBarView(contents []SideBarContent, width int) string {
	var b strings.Builder
	for _, content := range contents {
		b.WriteString(DocumentStyle.Highlight.Render(content.Title) + "\n\n")
		for _, key := range content.Order {
			var decoratedKey string
			if content.KeyHighlight {
				decoratedKey = DocumentStyle.Highlight.Render(key)
			} else {
				decoratedKey = DocumentStyle.Normal.Render(key)
			}

			var decoratedItem string
			if item, exists := content.Items[key]; exists {
				decoratedItem = DocumentStyle.Normal.Render(item)
			} else {
				decoratedItem = DocumentStyle.Normal.Render("--")
			}
			b.WriteString(fmt.Sprintf("%s: %s\n", decoratedKey, decoratedItem))
		}
		b.WriteString("\n")
	}

	return lipgloss.NewStyle().Width(width).Padding(0, 1).Render(b.String())
}

func MsgView(sizing ViewSize, msg string) string {
	return MsgStyle.Width(sizing.Width).
		Height(sizing.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(msg)
}

func ErrorView(sizing ViewSize, err error, withHelp bool) string {
	var msg string
	if err != nil {
		msg = ErrorStyle.Render("Error: " + err.Error())
	}

	errMsg := lipgloss.NewStyle().
		Width(sizing.Width).
		Height(sizing.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(msg)

	if withHelp {
		helper := HelperView([]HelperContent{{Key: "<", Action: "back"}}, sizing.Width, NormalView)
		return lipgloss.JoinVertical(
			lipgloss.Center,
			errMsg,
			helper,
		)
	}

	return errMsg
}

type ConfirmCheckItem struct {
	Label string // "Target", "Activity", "Session"
	Value string
}

type ConfirmCheck struct {
	Prompt  string
	Warning string
}

func (c ConfirmCheck) View(title string, sizing ViewSize) string {
	titleBar := TitleBarView(title, sizing.Width, false)
	helper := HelperView(
		[]HelperContent{{Key: "y", Action: "yes"}, {Key: "n", Action: "no"}},
		sizing.Width,
		NormalView,
	)

	var b strings.Builder
	b.WriteString(DocumentStyle.Highlight.Render(c.Prompt) + "\n\n")
	b.WriteString(WarningStyle.Render(c.Warning) + "\n")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleBar,
		lipgloss.NewStyle().
			Width(sizing.Width).
			Height(sizing.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(b.String()),
		helper,
	)
}
