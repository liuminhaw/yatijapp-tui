package main

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/model"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
	"github.com/liuminhaw/yatijapp-tui/internal/validator"
)

const (
	titleInputViewWidth   = 40
	dueDateInputViewWidth = 25
)

type targetPage struct {
	cfg    config
	mode   style.ViewMode
	action cmdAction

	uuid         string
	title        string
	fields       []Focusable
	focused      int
	focusedCache int

	width  int
	height int

	err  error
	prev tea.Model // Previous model for navigation
}

func newTargetPage(
	cfg config,
	title string,
	size style.ViewSize,
	target *data.Target,
	prev tea.Model,
) (targetPage, error) {
	focusables := []Focusable{}

	name := textinput.New()
	name.Focus()
	name.Prompt = ""
	name.Placeholder = "Give target a name"
	name.Width = titleInputViewWidth - 1
	name.CharLimit = 80
	name.Validate = validator.MultipleValidators(
		validator.ValidateRequired("title is required"),
		validator.ValidateReachMaxLength(80),
	)

	due := textinput.New()
	due.Placeholder = "YYYY-MM-DD"
	due.Width = dueDateInputViewWidth
	due.Prompt = ""
	due.Validate = validator.MultipleValidators(
		validator.ValidateDate(validator.ValidFormats),
		validator.ValidateDateAfter(time.Now().AddDate(0, 0, -1)),
	)

	description := textinput.New()
	description.Prompt = ""
	description.Placeholder = "Information about the target"
	description.Width = formWidth - 1
	description.CharLimit = 200
	description.Validate = validator.ValidateReachMaxLength(200)

	status := model.NewStatusModel([]string{"queued", "in progress", "completed", "canceled"})

	note := model.NewNoteModel()

	var uuid string
	targetAction := cmdCreate
	if target != nil {
		targetAction = cmdUpdate

		name.SetValue(target.Title)
		if target.DueDate.Valid {
			due.SetValue(target.DueDate.Time.Format("2006-01-02"))
		}
		description.SetValue(target.Description)

		if err := status.SetValue(target.Status); err != nil {
			return targetPage{}, internalErrorMsg{msg: "failed to load target data", err: err}
		}
		if err := note.SetValue(target.Notes); err != nil {
			return targetPage{}, internalErrorMsg{msg: "failed to load target data", err: err}
		}

		uuid = target.UUID
	}
	focusables = append(
		focusables,
		model.NewTextInputWrapper(name),
		model.NewTextInputWrapper(due),
		model.NewTextInputWrapper(description),
		status,
		note,
	)

	return targetPage{
		cfg:    cfg,
		mode:   style.NormalView,
		action: targetAction,
		uuid:   uuid,
		title:  title,
		fields: focusables,
		width:  size.Width,
		height: size.Height,
		prev:   prev,
	}, nil
}

func (m targetPage) Init() tea.Cmd {
	return nil
}

func (m targetPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case data.LoadApiDataErr:
		m.cfg.logger.Error(msg.Error(), slog.Int("status", msg.Status), slog.String("action", "save target"))
		m.err = errors.New(msg.Msg)
	case tea.KeyMsg:
		switch m.mode {
		case style.NormalView:
			switch msg.String() {
			case "tab", "enter":
				// Cycle through focusable fields
				m.fields[m.focused].Validate()
				m.fields[m.focused].Blur()
				m.focused = (m.focused + 1) % len(m.fields)
				m.focusedCache = m.focused
				return m, m.fields[m.focused].Focus()
			case "shift+tab":
				m.fields[m.focused].Validate()
				m.fields[m.focused].Blur()
				m.focused = (m.focused - 1 + len(m.fields)) % len(m.fields)
				m.focusedCache = m.focused
				return m, m.fields[m.focused].Focus()
			case "esc", "ctrl+[":
				m.mode = style.HighlightView
				m.fields[m.focused].Blur()
				m.focused = -1
			}
		case style.HighlightView:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "<":
				return m, switchToPreviousCmd(m.prev)
			case "ctrl+s":
				switch m.action {
				case cmdCreate:
					return m, m.create()
				case cmdUpdate:
					return m, m.update()
				}
			case "e":
				m.mode = style.NormalView
				m.focused = m.focusedCache
				m.fields[m.focused].Focus()
				return m, nil
			}
		}
	// unexpectedApiResponseMsg, validationErrorMsg
	case error:
		m.err = msg
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

func (m targetPage) View() string {
	name := m.fields[0]
	due := m.fields[1]
	description := m.fields[2]
	status := m.fields[3]
	note := m.fields[4]

	titleView := style.TitleBarView(m.title, viewWidth, false)

	var notePrompt string
	if note.Focused() {
		notePrompt = fmt.Sprintf(
			"%s %s\n%s\n",
			style.FormFieldStyle.Prompt("Note", note.Focused()),
			style.FormFieldStyle.Helper.Render("(Press 'e' to edit)"),
			style.FormFieldStyle.Error.Render(note.Error()),
		)
	} else {
		notePrompt = fmt.Sprintf(
			"%s\n%s\n",
			style.FormFieldStyle.Prompt("Note", note.Focused()),
			style.FormFieldStyle.Error.Render(note.Error()),
		)
	}

	configContent := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().
				Width(titleInputViewWidth).Margin(1, 2, 0, 5).Render(
				fmt.Sprintf(
					"%s\n%s\n%s",
					style.FormFieldStyle.Prompt("Name", name.Focused()),
					style.FormFieldStyle.Content.Render(name.View()),
					style.FormFieldStyle.Error.Render(name.Error()),
				),
			),
			lipgloss.NewStyle().
				Width(dueDateInputViewWidth).Margin(1, 0, 0, 3).Render(
				fmt.Sprintf(
					"%s\n%s\n%s",
					style.FormFieldStyle.Prompt("Due Date", due.Focused()),
					style.FormFieldStyle.Content.Render(due.View()),
					style.FormFieldStyle.Error.Render(due.Error()),
				),
			),
		),

		lipgloss.NewStyle().Width(formWidth).Margin(1, 5, 0).Render(
			fmt.Sprintf(
				"%s\n%s\n%s",
				style.FormFieldStyle.Prompt("Description", description.Focused()),
				style.FormFieldStyle.Content.Render(description.View()),
				style.FormFieldStyle.Error.Render(description.Error()),
			),
		),

		lipgloss.NewStyle().Width(formWidth).Margin(1, 5, 0).Render(
			fmt.Sprintf(
				"%s\n%s\n",
				style.FormFieldStyle.Prompt("Status", status.Focused()),
				style.FormFieldStyle.Content.Render(status.View()),
			),
		),

		lipgloss.NewStyle().Width(formWidth).Margin(1, 5, 0).Render(notePrompt),
	)

	var helperContent []style.HelperContent
	switch m.mode {
	case style.NormalView:
		helperContent = []style.HelperContent{
			{Key: "Esc", Action: "finish editing"},
			{Key: "Tab/Shift+Tab", Action: "navigate"},
		}
	case style.HighlightView:
		var saveAction string
		if m.action == cmdCreate {
			saveAction = "create"
		} else {
			saveAction = "save"
		}

		helperContent = []style.HelperContent{
			{Key: "<", Action: "back"},
			{Key: "e", Action: "edit mode"},
			{Key: "<C-s>", Action: saveAction},
			{Key: "q", Action: "quit"},
		}
	}

	helperView := style.HelperView(helperContent, viewWidth, m.mode)

	msgView := style.ErrorView(style.ViewSize{Width: viewWidth, Height: 1}, m.err, nil)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		titleView,
		configContent,
		msgView,
		helperView,
	)

	return style.ContainerStyle(m.width, container, 5).Render(container)
}

func (m targetPage) validationError() error {
	for _, field := range m.fields {
		field.Validate()
		if err := field.Error(); err != "" {
			return errors.New("Input validation failed")
		}
	}

	return nil
}

func (m targetPage) create() tea.Cmd {
	title := m.fields[0].Value()
	due := m.fields[1].Value()
	description := m.fields[2].Value()
	status := m.fields[3].Value()
	note := m.fields[4].Value()

	if err := m.validationError(); err != nil {
		return validationErrorCmd(err)
	}

	request := data.TargetRequestBody{
		Title:       title,
		Description: description,
		Status:      status,
		Notes:       note,
	}
	if due != "" {
		dueDate, err := time.ParseInLocation("2006-01-02", due, time.Local)
		if err != nil {
			return validationErrorCmd(errors.New("invalid due date format"))
		}
		dueDateTyped := data.Date(dueDate)
		request.DueDate = &dueDateTyped
	}

	if err := request.Create(m.cfg.serverURL, m.cfg.authClient); err != nil {
		return func() tea.Msg { return apiErrorResponseCmd(err) }
	}

	return apiSuccessResponseCmd("Target created successfully", m.prev)
}

func (m targetPage) update() tea.Cmd {
	title := m.fields[0].Value()
	due := m.fields[1].Value()
	description := m.fields[2].Value()
	status := m.fields[3].Value()
	note := m.fields[4].Value()

	if err := m.validationError(); err != nil {
		return validationErrorCmd(err)
	}

	request := data.TargetRequestBody{
		Title:       title,
		Description: description,
		Status:      status,
		Notes:       note,
	}
	if due != "" {
		dueDate, err := time.ParseInLocation("2006-01-02", due, time.Local)
		if err != nil {
			return validationErrorCmd(errors.New("invalid due date format"))
		}
		dueDateTyped := data.Date(dueDate)
		request.DueDate = &dueDateTyped
	}

	if err := request.Update(m.cfg.serverURL, m.uuid, m.cfg.authClient); err != nil {
		return func() tea.Msg { return apiErrorResponseCmd(err) }
	}

	return apiSuccessResponseCmd("Target updated successfully", m.prev)
}
