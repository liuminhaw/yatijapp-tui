package model

import (
	"errors"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TextInputWrapper struct {
	model textinput.Model
}

func NewTextInputWrapper(input textinput.Model) *TextInputWrapper {
	return &TextInputWrapper{
		model: input,
	}
}

func (t *TextInputWrapper) Focus() tea.Cmd {
	return t.model.Focus()
}

func (t *TextInputWrapper) Focused() bool {
	return t.model.Focused()
}

func (t *TextInputWrapper) Blur() {
	t.model.Blur()
}

func (t *TextInputWrapper) Init() tea.Cmd {
	return nil
}

func (t *TextInputWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	return t, cmd
}

func (t *TextInputWrapper) View() string {
	return t.model.View()
}

func (t *TextInputWrapper) Value() string {
	return t.model.Value()
}

func (t *TextInputWrapper) Values() []string {
	return []string{t.model.Value()}
}

func (t *TextInputWrapper) SetValues(vals ...string) error {
	if len(vals) != 1 {
		return errors.New("TextInputWrapper expects a single value")
	}
	t.model.SetValue(vals[0])

	return nil
}

func (t *TextInputWrapper) Validate() {
	if t.model.Validate != nil {
		t.model.Err = t.model.Validate(t.model.Value())
	}
}

func (t *TextInputWrapper) Error() string {
	if t.model.Err == nil {
		return ""
	}
	return t.model.Err.Error()
}

func (t *TextInputWrapper) Clear() {
	t.model.SetValue("")
	t.model.Err = nil
}

func ValidateTextInputMatch(target *TextInputWrapper, msg string) textinput.ValidateFunc {
	return func(input string) error {
		if input != target.Value() {
			return errors.New(msg)
		}
		return nil
	}
}
