package model

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/components"
)

var (
	StatusOptions        = []string{"queued", "in progress", "completed", "canceled"}
	SortByOptions        = []string{"id", "due date", "created at", "last active"}
	SessionStatusOptions = []string{"in progress", "completed"}
	SessionSortByOptions = []string{"starts at", "updated at"}
	SortOrderOptions     = []string{"ascending", "descending"}
)

type RadioModel struct {
	model components.Radio

	width int
	err   error
}

func NewRadioModel(choices []string, width int) *RadioModel {
	// func NewStatusModel(width int) *RadioModel {
	return &RadioModel{
		model: components.NewRadio(choices, components.RadioDefaultView, width),
	}
}

func (m *RadioModel) Init() tea.Cmd {
	return nil
}

func (m *RadioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)
	return m, cmd
}

func (m *RadioModel) View() string {
	return m.model.View()
}

func (m *RadioModel) Value() string {
	return m.model.Selected()
}

func (m *RadioModel) Values() []string {
	return []string{m.model.Selected()}
}

func (m *RadioModel) Focus() tea.Cmd {
	return m.model.Focus()
}

func (m *RadioModel) Focused() bool {
	return m.model.Focused()
}

func (m *RadioModel) Blur() {
	m.model.Blur()
}

func (m *RadioModel) Validate() {}
func (m *RadioModel) Error() string {
	if m.err != nil {
		return m.err.Error()
	}
	return ""
}

func (m *RadioModel) SetValues(vals ...string) error {
	if len(vals) != 1 {
		return errors.New("StatusModel expects a single value")
	}

	if err := m.model.SetValue(vals[0]); err != nil {
		return err
	}

	return nil
}

type CheckboxModel struct {
	model components.Checkbox

	width int
	err   error
}

func NewCheckboxModel(choices []string, width int) *CheckboxModel {
	return &CheckboxModel{
		model: components.NewCheckbox(choices, width),
	}
}

func (m *CheckboxModel) Init() tea.Cmd {
	return nil
}

func (m *CheckboxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.model, cmd = m.model.Update(msg)
	return m, cmd
}

func (m *CheckboxModel) View() string {
	return m.model.View()
}

func (m *CheckboxModel) Value() string {
	return ""
}

func (m *CheckboxModel) Values() []string {
	return m.model.Values()
}

func (m *CheckboxModel) Focus() tea.Cmd {
	return m.model.Focus()
}

func (m *CheckboxModel) Focused() bool {
	return m.model.Focused()
}

func (m *CheckboxModel) Blur() {
	m.model.Blur()
}

func (m *CheckboxModel) Validate() {}
func (m *CheckboxModel) Error() string {
	if m.err != nil {
		return m.err.Error()
	}
	return ""
}

func (m *CheckboxModel) SetValues(vals ...string) error {
	return m.model.SetValues(vals...)
}
