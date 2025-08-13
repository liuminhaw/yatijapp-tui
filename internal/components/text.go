package components

import tea "github.com/charmbracelet/bubbletea"

type TextComponent struct {
	text    string
	focused bool
	err     error
}

func NewText(text string) *TextComponent {
	return &TextComponent{
		text:    text,
		focused: false,
	}
}

func (t *TextComponent) Blur() {
	t.focused = false
}

func (t *TextComponent) Focus() tea.Cmd {
	t.focused = true
	return nil
}

func (t *TextComponent) Focused() bool {
	return t.focused
}

func (t *TextComponent) Init() tea.Cmd {
	return nil
}

func (t *TextComponent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return t, nil
}

func (t *TextComponent) View() string {
	return t.text
}

func (t *TextComponent) Value() string {
	return t.text
}

func (t *TextComponent) Validate() {}
func (t *TextComponent) Error() string {
	if t.err != nil {
		return t.err.Error()
	}
	return ""
}
