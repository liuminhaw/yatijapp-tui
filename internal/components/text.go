package components

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

type TextComponent struct {
	text    string
	focused bool

	err error

	msg          tea.Msg
	ValidateFunc func(string) error
}

func NewText(text string, msg tea.Msg) *TextComponent {
	return &TextComponent{
		text:    text,
		focused: false,
		msg:     msg,
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			if !t.focused {
				return t, nil
			}
			return t, func() tea.Msg { return t.msg }
		}
	}

	return t, nil
}

func (t *TextComponent) View() string {
	return t.text
}

func (t *TextComponent) Value() string {
	return t.text
}

func (t *TextComponent) Values() []string {
	return []string{t.text}
}

func (t *TextComponent) SetValues(vals ...string) error {
	if len(vals) != 1 {
		return errors.New("TextComponent expects a single value")
	}
	t.text = vals[0]

	return nil
}

func (t *TextComponent) Validate() {
	if t.ValidateFunc != nil {
		t.err = t.ValidateFunc(t.text)
	}
}

func (t *TextComponent) Error() string {
	if t.err != nil {
		return t.err.Error()
	}
	return ""
}
