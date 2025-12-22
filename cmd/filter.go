package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/model"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type filterPage struct {
	cfg config

	recordType data.RecordType
	fields     []Focusable
	focused    int

	width  int
	height int

	err  error
	prev tea.Model
}

func newFilterPage(
	cfg config,
	f data.RecordFilter,
	size style.ViewSize,
	prev tea.Model,
) filterPage {
	focusables := []Focusable{}

	sortOrder := model.NewRadioModel(model.SortOrderOptions, formWidth)

	var sortBy, status Focusable
	if f.RecordType == data.RecordTypeSession {
		sortBy = model.NewRadioModel(model.SessionSortByOptions, formWidth)
		status = model.NewCheckboxModel(model.SessionStatusOptions, formWidth)
	} else {
		sortBy = model.NewRadioModel(model.SortByOptions, formWidth)
		status = model.NewCheckboxModel(model.StatusOptions, formWidth)
	}

	cfg.logger.Info("filter", "filter", fmt.Sprintf("%+v", f.Filter))
	if err := sortBy.SetValues(f.Filter.SortOption()); err != nil {
		panic("failed to set sortBy values: " + err.Error())
	}
	if err := sortOrder.SetValues(f.Filter.SortOrder); err != nil {
		panic("failed to set sortOrder values: " + err.Error())
	}
	if err := status.SetValues(f.Filter.Status...); err != nil {
		panic("failed to set status values: " + err.Error())
	}

	focusables = append(focusables, sortBy, sortOrder, status)
	focused := 0
	focusables[focused].Focus()

	return filterPage{
		cfg:        cfg,
		recordType: f.RecordType,
		fields:     focusables,
		width:      size.Width,
		height:     size.Height,
		prev:       prev,
	}
}

func (m filterPage) Init() tea.Cmd {
	return nil
}

func (m filterPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "enter":
			// Cycle through focusable fields
			m.fields[m.focused].Validate()
			m.fields[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.fields)
			// m.focusedCache = m.focused
			return m, m.fields[m.focused].Focus()
		case "shift+tab":
			m.fields[m.focused].Validate()
			m.fields[m.focused].Blur()
			m.focused = (m.focused - 1 + len(m.fields)) % len(m.fields)
			// m.focusedCache = m.focused
			return m, m.fields[m.focused].Focus()
		case "ctrl+s":
			return m, switchToPreviousCmd(m.prevPage())
		case "esc", "ctrl+[":
			return m, switchToPreviousCmd(m.prev)
		}
		if isFnKey(msg) {
			n, err := fnKeyNumber(msg)
			if err != nil {
				panic("failed to get function key number: " + err.Error())
			}
			if n >= 1 && n <= len(m.fields) {
				m.fields[m.focused].Validate()
				m.fields[m.focused].Blur()
				m.focused = n - 1
				// m.focusedCache = m.focused
				return m, m.fields[m.focused].Focus()
			}
		}
	}

	for i, field := range m.fields {
		retModel, retCmd := field.Update(msg)
		m.fields[i] = retModel.(Focusable)
		if retCmd != nil {
			cmds = append(cmds, retCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m filterPage) View() string {
	sortBy := field{idx: 0, obj: m.fields[0]}
	sortOrder := field{idx: 1, obj: m.fields[1]}
	status := field{idx: 2, obj: m.fields[2]}

	titleView := style.TitleBarView([]string{"Filter"}, viewWidth, false)

	statusPrompt := status.prompt(
		status.simpleTitlePrompt(
			"Status",
			"(←/→ keys to select, space to toggle selection)",
			false,
		),
		status.simpleTitlePrompt("Status", fmt.Sprintf("<f%d>", status.idx+1), false),
	)

	filterForm := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Width(formWidth).Margin(1, 5, 0).Render(
			fmt.Sprintf(
				"%s\n%s\n",
				sortBy.prompt(
					sortBy.simpleTitlePrompt("Sort By", "(Use ←/→ keys to select)", false),
					sortBy.simpleTitlePrompt("Sort By", fmt.Sprintf("<f%d>", sortBy.idx+1), false),
				),
				style.FormFieldStyle.Content.Render(sortBy.obj.View()),
			),
		),

		style.BorderStyle["normal"].BorderBottom(true).
			Width(formWidth).
			Margin(1, 5, 0).
			Render(
				fmt.Sprintf(
					"%s\n%s\n",
					sortOrder.prompt(
						sortOrder.simpleTitlePrompt(
							"Sort Order",
							"(Use ←/→ keys to select)",
							false,
						),
						sortOrder.simpleTitlePrompt(
							"Sort Order",
							fmt.Sprintf("<f%d>", sortOrder.idx+1),
							false,
						),
					),
					style.FormFieldStyle.Content.Render(sortOrder.obj.View()),
				),
			),

		lipgloss.NewStyle().Width(formWidth).Margin(1, 5, 0).Render(
			fmt.Sprintf(
				"%s\n%s\n",
				statusPrompt,
				style.FormFieldStyle.Content.Render(status.obj.View()),
			),
		),
	)

	helperContent := []style.HelperContent{
		{Key: "Esc", Action: "back"},
		{Key: "Tab/Shift+Tab", Action: "navigate"},
		{Key: "<C-s>", Action: "submit"},
		{Key: "<C-c>", Action: "quit"},
	}

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		titleView,
		filterForm,
		style.HelperView(helperContent, viewWidth),
	)

	return style.ContainerStyle(m.width, container, 5).Render(container)
}

func (m filterPage) getFilter() data.RecordFilter {
	return data.RecordFilter{
		RecordType: m.recordType,
		Filter: data.Filter{
			SortBy:    m.fields[0].Value(),
			SortOrder: m.fields[1].Value(),
			Status:    m.fields[2].Values(),
		},
	}
}

func (m filterPage) prevPage() tea.Model {
	filter := m.getFilter()

	if v, ok := m.prev.(listPage); ok {
		v.selectionFilterQuery(filter)
		m.cfg.logger.Info("switch to previous page", "prev", fmt.Sprintf("%+v", v.filter))
		return v
	}

	if v, ok := m.prev.(menuPage); ok {
		switch filter.RecordType {
		case data.RecordTypeTarget:
			v.cfg.preferences.Filters.Target = filter.Filter
		case data.RecordTypeAction:
			v.cfg.preferences.Filters.Action = filter.Filter
		case data.RecordTypeSession:
			v.cfg.preferences.Filters.Session = filter.Filter
		}

		request := data.NewPreferencesRequestBody(*v.cfg.preferences)
		if err := request.Update(m.cfg.apiEndpoint, m.cfg.authClient); err != nil {
			v.error = fmt.Errorf("failed to update preferences: %w", err)
			return v
		}

		return v
	}

	return m.prev
}
