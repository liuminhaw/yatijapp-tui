package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
)

type ButtonPressedMsg struct {
	Label string
}

type Button struct {
	label   string
	focused bool
}

func NewButton(label string) Button {
	return Button{
		label:   label,
		focused: false,
	}
}

func (b *Button) Blur() {
	b.focused = false
}

func (b *Button) Focus() tea.Cmd {
	b.focused = true
	return nil
}

func (b *Button) Focused() bool {
	return b.focused
}

func (b *Button) Init() tea.Cmd {
	return nil
}

func (b *Button) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !b.focused {
		return b, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return b, func() tea.Msg {
				return ButtonPressedMsg{Label: b.label}
			}
		}
	}

	return b, nil
}

func (b *Button) View() string {
	if b.focused {
		return b.focusedView()
	} else {
		return b.blurredView()
	}
}

func (b *Button) Value() string {
	return b.label
}

func (b *Button) Validate() {}
func (b *Button) Error() string {
	return ""
}

func (b *Button) focusedView() string {
	return lipgloss.NewStyle().
		Foreground(colors.DocumentTextBright).Render("[" + b.label + "]")
}

func (b *Button) blurredView() string {
	return lipgloss.NewStyle().Foreground(colors.DocumentText).Render("[ ") +
		lipgloss.NewStyle().Foreground(colors.DocumentTextDim).Render(b.label) +
		lipgloss.NewStyle().Foreground(colors.DocumentText).Render(" ]")
}
