package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
	"github.com/liuminhaw/yatijapp-tui/internal/validator"
)

type searchPage struct {
	cfg config

	scope data.RecordType
	field Focusable

	width  int
	height int
	prev   tea.Model
}

func newSearchPage(
	cfg config,
	scope data.RecordType,
	size style.ViewSize,
	content string,
	prev tea.Model,
) searchPage {
	fieldWidth := formWidth - 2
	field := generalInput(inputFieldConfig{
		width:       fieldWidth,
		focus:       true,
		placeholder: "Search...",
		lenMax:      fieldWidth - 1,
		validators: []func(string) error{
			validator.ValidateRequired("required"),
			validator.ValidateMaxLength(fieldWidth - 1),
		},
	})

	if content != "" {
		field.SetValues(content)
	}

	return searchPage{
		cfg:    cfg,
		scope:  scope,
		field:  field,
		width:  size.Width,
		height: size.Height,
		prev:   prev,
	}
}

func (s searchPage) Init() tea.Cmd {
	return nil
}

func (s searchPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return s, tea.Quit
		case "esc", "ctrl+[":
			return s, switchToPreviousCmd(s.prev)
		case "enter":
			if s.scope == data.RecordTypeAll {
				return s, switchToSearchListCmd(s.field.Value())
			}
			return s, switchToPreviousCmd(s.prevPage())
		}
	}

	retModel, retCmd := s.field.Update(msg)
	s.field = retModel.(Focusable)

	return s, retCmd
}

func (s searchPage) View() string {
	search := field{obj: s.field}
	title := ""

	switch s.scope {
	case data.RecordTypeAll:
		title = "Search All"
	case data.RecordTypeTarget, data.RecordTypeAction, data.RecordTypeSession:
		title = "Search " + string(s.scope) + "s"
	default:
		panic("unknown search scope: " + s.scope)
	}

	helper := lipgloss.NewStyle().Width(formWidth).AlignHorizontal(lipgloss.Center).Render(
		lipgloss.StyleRanges(
			"Esc back    Enter search",
			lipgloss.Range{Start: 0, End: 3, Style: style.Document.Primary},
			lipgloss.Range{Start: 4, End: 8, Style: style.Document.Normal},
			lipgloss.Range{Start: 12, End: 17, Style: style.Document.Primary},
			lipgloss.Range{Start: 18, End: 24, Style: style.Document.Normal},
		),
	)

	title = style.InputStyle.Selected.Width(formWidth).
		AlignHorizontal(lipgloss.Center).
		Margin(0, 0, 1).
		Render(title)

	form := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		style.FormFieldStyle.Content.
			Width(formWidth-2).
			Render(search.obj.View()),
		helper,
	)

	return style.BorderStyle["highlighted"].Width(formWidth).Render(form)
}

func (s searchPage) prevPage() tea.Model {
	search := s.field.Value()

	if v, ok := s.prev.(listPage); ok {
		v.selectionSearchQuery(search)
		return v
	}

	panic("previous page is not listPage")
}
