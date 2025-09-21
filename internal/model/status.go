package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/components"
)

type StatusModel struct {
	model components.Radio

	width int
	err   error
}

func NewStatusModel(choices []string, width int) *StatusModel {
	return &StatusModel{
		model: components.NewRadio(choices, components.RadioDefaultView, width),
	}
}

func (m *StatusModel) Init() tea.Cmd {
	return nil
}

func (m *StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)
	return m, cmd
}

func (m *StatusModel) View() string {
	return m.model.View()
}

func (m *StatusModel) Value() string {
	return m.model.Selected()
	// return m.model.Value()
}

func (m *StatusModel) Focus() tea.Cmd {
	return m.model.Focus()
}

func (m *StatusModel) Focused() bool {
	return m.model.Focused()
}

func (m *StatusModel) Blur() {
	m.model.Blur()
}

func (m *StatusModel) Validate() {}
func (m *StatusModel) Error() string {
	if m.err != nil {
		return m.err.Error()
	}
	return ""
}

func (m *StatusModel) SetValue(value string) error {
	if err := m.model.SetValue(value); err != nil {
		return err
	}

	return nil
}
