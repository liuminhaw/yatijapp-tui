package style

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
)

type ViewMode int

type ViewSize struct {
	Width  int
	Height int
}

func TitleBarView(contents []string, width int, msg bool) string {
	title := Document.Highlight.Render("Yatijapp")
	if len(contents) == 0 {
		return title
	}

	if msg {
		for _, content := range contents {
			title += Document.Normal.Render(" - ")
			title += MsgStyle.Render(content)
		}
	} else {
		if len(contents) == 1 {
			title += Document.NormalDim.Render(" - " + contents[0])
		} else {
			content := strings.Join(contents[:len(contents)-1], " - ")
			title += Document.NormalDim.Render(" - " + content)
			title += Document.Normal.Render(" - " + contents[len(contents)-1])
		}
	}

	return BorderStyling.Width(width).Padding(0, 1).Render(title)
}

func LoadingView(s *spinner.Model, title string, sizing ViewSize) string {
	titleBar := TitleBarView([]string{title}, sizing.Width, false)
	helper := HelperView([]HelperContent{{Key: "<", Action: "back"}}, sizing.Width)

	msg := Document.NormalDim.Bold(true).Render("loading...")
	s.Style = Document.Highlight

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

func HelperView(contents []HelperContent, width int) string {
	var b strings.Builder
	for i, content := range contents {
		b.WriteString(
			fmt.Sprintf(
				"%s %s",
				HelperStyle.Key.Render(content.Key),
				HelperStyle.Action.Render(content.Action),
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

type FullHelpContent struct {
	Title        string
	Items        map[string]string
	Order        []string // Order of items to maintain display order
	KeyHighlight bool
}

func FullHelpView(contents []FullHelpContent, width int) string {
	var b strings.Builder
	for _, content := range contents {
		b.WriteString(Document.Primary.Bold(true).Render(content.Title) + "\n\n")
		for i, key := range content.Order {
			var decoratedKey string
			if content.KeyHighlight {
				decoratedKey = Document.Highlight.Render(key)
			} else {
				decoratedKey = Document.NormalDim.Render(key)
			}

			var decoratedItem string
			if item, exists := content.Items[key]; exists {
				decoratedItem = Document.NormalDim.Render(item)
			} else {
				decoratedItem = Document.NormalDim.Render("--")
			}
			b.WriteString(decoratedKey + Document.NormalDim.Render(": ") + decoratedItem)
			if i < len(content.Order)-1 {
				b.WriteString("\n")
			}
		}
	}

	return lipgloss.NewStyle().
		Width(width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colors.Text).
		Padding(0, 1).
		Render(b.String())
}

func MsgView(sizing ViewSize, msg string) string {
	return MsgStyle.Width(sizing.Width).
		Height(sizing.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(msg)
}

func ErrorView(sizing ViewSize, err error, helpMsg []HelperContent) string {
	var msg string
	if err != nil {
		msg = ErrorStyle.Render("Error: " + err.Error())
	}

	errMsg := lipgloss.NewStyle().
		Width(sizing.Width).
		Height(sizing.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(msg)

	if helpMsg != nil {
		helper := HelperView(helpMsg, sizing.Width)
		return lipgloss.JoinVertical(
			lipgloss.Center,
			errMsg,
			helper,
		)
	}

	return errMsg
}

func FullPageErrorView(
	title string,
	termWidth int,
	sizing ViewSize,
	err error,
	helpMsg []HelperContent,
) string {
	container := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		ErrorView(sizing, err, helpMsg),
	)

	return ContainerStyle(termWidth, container, 5).Render(container)
}

type ConfirmCheckItem struct {
	Label string // "Target", "Action", "Session"
	Value string
}

type ConfirmCheck struct {
	Prompt  string
	Warning string
	Cmd     tea.Cmd
}

func (c ConfirmCheck) View(title string, width int) string {
	var b strings.Builder
	b.WriteString(Document.Secondary.Bold(true).Render(title) + "\n\n")
	b.WriteString(Document.Highlight.Render(c.Prompt) + "\n")
	b.WriteString(ErrorStyle.Render(c.Warning) + "\n\n")
	b.WriteString(lipgloss.StyleRanges(
		"[y]es        [n]o",
		lipgloss.Range{Start: 0, End: 3, Style: Document.Primary},
		lipgloss.Range{Start: 3, End: 5, Style: Document.Normal},
		lipgloss.Range{Start: 13, End: 16, Style: Document.Primary},
		lipgloss.Range{Start: 16, End: 17, Style: Document.Normal},
	))

	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().
			Width(width).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.Text).
			Padding(0, 1).
			Render(b.String()),
	)
}

func NewPaginator(perPage int) paginator.Model {
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = perPage
	p.ActiveDot = lipgloss.NewStyle().Foreground(colors.DocumentText).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(colors.HelperTextDim).Render("•")

	return p
}
