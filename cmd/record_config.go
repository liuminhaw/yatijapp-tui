package main

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
	"github.com/liuminhaw/yatijapp-tui/internal/components"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/model"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
	"github.com/liuminhaw/yatijapp-tui/internal/validator"
	"github.com/liuminhaw/yatijapp-tui/pkg/strview"
)

const (
	titleInputViewWidth   = 51
	dueDateInputViewWidth = 14
)

type recordConfigHooks struct {
	create func(
		serverURL string,
		body recordRequestData,
		src, redirect tea.Model,
		client *authclient.AuthClient,
	) tea.Cmd
	update func(
		serverURL string,
		body recordRequestData,
		src, redirect tea.Model,
		client *authclient.AuthClient,
	) tea.Cmd
}

type recordConfigPage struct {
	cfg    config
	mode   style.ViewMode
	action cmdAction
	// createWithSrc bool

	uuid       string
	record     yatijappRecord
	recordType data.RecordType
	hooks      recordConfigHooks

	title        string
	fields       []Focusable
	focused      int
	focusedCache int
	// hiddenFields map[string]struct{ val string }
	hiddenFields map[string]string

	width  int
	height int

	err  error
	prev tea.Model // Previous model for navigation

	selector tea.Model
}

func newRecordConfigPage(
	cfg config,
	title string,
	size style.ViewSize,
	record yatijappRecord,
	recordType data.RecordType,
	// createWithSrc bool,
	prev tea.Model,
) (recordConfigPage, error) {
	focusables := []Focusable{}
	// hiddens := make(map[string]struct{ val string })
	hiddens := make(map[string]string)

	name := textinput.New()
	name.Prompt = ""
	name.Placeholder = "Give " + strings.ToLower(string(recordType)) + " a name"
	name.Width = titleInputViewWidth - 1
	name.CharLimit = 80
	name.Validate = validator.MultipleValidators(
		validator.ValidateRequired("title is required"),
		validator.ValidateReachMaxLength(80),
	)

	due := textinput.New()
	due.Placeholder = "YYYY-MM-DD"
	due.Width = dueDateInputViewWidth - 1
	due.Prompt = ""
	due.Validate = validator.MultipleValidators(
		validator.ValidateDate(validator.ValidFormats),
		validator.ValidateDateAfter(time.Now().AddDate(0, 0, -1)),
	)

	description := textinput.New()
	description.Prompt = ""
	description.Placeholder = "Information about the " + strings.ToLower(string(recordType))
	description.Width = formWidth - 1
	description.CharLimit = 200
	description.Validate = validator.ValidateReachMaxLength(200)

	status := model.NewStatusModel(
		[]string{"queued", "in progress", "completed", "canceled"},
		formWidth,
	)

	note := model.NewNoteModel()

	// For actions, and sessions
	parentTarget := components.NewText("", showSelectorMsg{})
	parentTarget.ValidateFunc = validator.ValidateRequired("target is required")

	var uuid string
	recordAction := cmdCreate
	if record != nil {
		if record.GetTitle() != "" {
			recordAction = cmdUpdate
			name.SetValue(record.GetTitle())
		}

		dueDate, dueDateValid := record.GetDueDate()
		if dueDateValid {
			due.SetValue(dueDate.Format("2006-01-02"))
		}
		description.SetValue(record.GetDescription())

		if err := status.SetValue(record.GetStatus()); err != nil {
			return recordConfigPage{}, internalErrorMsg{
				msg: "failed to load " + strings.ToLower(string(recordType)) + " status data",
				err: err,
			}
		}
		if err := note.SetValue(record.GetNote()); err != nil {
			return recordConfigPage{}, internalErrorMsg{
				msg: "failed to load " + strings.ToLower(string(recordType)) + " note data",
				err: err,
			}
		}

		parentTarget.SetValue(record.GetParentsTitle()[data.RecordTypeTarget])
		// hiddens["parent_target_uuid"] = struct{ val string }{
		// 	val: record.GetParentsUUID()[data.RecordTypeTarget],
		// }
		hiddens["parent_target_uuid"] = record.GetParentsUUID()[data.RecordTypeTarget]

		uuid = record.GetUUID()
	}
	focusables = append(
		focusables,
		model.NewTextInputWrapper(name),
		model.NewTextInputWrapper(due),
		model.NewTextInputWrapper(description),
		status,
		note,
	)

	switch recordType {
	case data.RecordTypeTarget:
		focusables[0].Focus() // Focus name field
	case data.RecordTypeAction:
		focusables = append(focusables, parentTarget)
		if record == nil {
			focusables[5].Focus() // Focus parent target field
		} else {
			focusables[0].Focus() // Focus name field
		}
	}

	return recordConfigPage{
		cfg:          cfg,
		mode:         style.NormalView,
		action:       recordAction,
		record:       record,
		recordType:   recordType,
		uuid:         uuid,
		title:        title,
		fields:       focusables,
		hiddenFields: hiddens,
		width:        size.Width,
		height:       size.Height,
		prev:         prev,
	}, nil
}

func newTargetConfigPage(
	cfg config,
	title string,
	size style.ViewSize,
	record yatijappRecord,
	prev tea.Model,
) (recordConfigPage, error) {
	page, err := newRecordConfigPage(
		cfg, title, size, record, data.RecordTypeTarget, prev,
	)
	if err != nil {
		return recordConfigPage{}, err
	}
	page.hooks = recordConfigHooks{
		create: createTarget,
		update: updateTarget,
	}

	return page, nil
}

func newActionConfigPage(
	cfg config,
	title string,
	size style.ViewSize,
	record yatijappRecord,
	// hasSrc bool,
	prev tea.Model,
) (recordConfigPage, error) {
	page, err := newRecordConfigPage(
		cfg, title, size, record, data.RecordTypeAction, prev,
	)
	if err != nil {
		return recordConfigPage{}, err
	}
	page.hooks = recordConfigHooks{
		create: createAction,
		update: updateAction,
	}
	if record == nil {
		page.focused = len(page.fields) - 1
	}

	return page, nil
}

func (p recordConfigPage) Init() tea.Cmd {
	return nil
}

func (p recordConfigPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
	case tea.KeyMsg:
		if p.selector != nil {
			break
		}
		switch p.mode {
		case style.NormalView:
			switch msg.String() {
			case "tab", "enter":
				// Cycle through focusable fields
				p.fields[p.focused].Validate()
				p.fields[p.focused].Blur()
				p.focused = (p.focused + 1) % len(p.fields)
				p.focusedCache = p.focused
				return p, p.fields[p.focused].Focus()
			case "shift+tab":
				p.fields[p.focused].Validate()
				p.fields[p.focused].Blur()
				p.focused = (p.focused - 1 + len(p.fields)) % len(p.fields)
				p.focusedCache = p.focused
				return p, p.fields[p.focused].Focus()
			case "esc", "ctrl+[":
				p.mode = style.HighlightView
				p.fields[p.focused].Blur()
				p.focused = -1
			}
		case style.HighlightView:
			switch msg.String() {
			case "ctrl+c", "q":
				return p, tea.Quit
			case "<":
				return p, switchToPreviousCmd(p.prev)
			case "ctrl+s":
				switch p.action {
				case cmdCreate:
					return p, p.create()
				case cmdUpdate:
					p.cfg.logger.Info("updating record", slog.String("uuid", p.uuid))
					return p, p.update()
				}
			case "e":
				p.mode = style.NormalView
				p.focused = p.focusedCache
				p.fields[p.focused].Focus()
				return p, nil
			}
		}
	case showSelectorMsg:
		p.selector = newTargetSelectorPage(p.cfg, style.ViewSize{Width: p.width, Height: p.height}, p)
		cmd := p.selector.Init()
		cmds = append(cmds, cmd)
	case selectorTargetSelectedMsg:
		p.selector = nil
		if len(p.fields) < 6 {
			panic("action config page missing target field")
		}
		p.fields[5].SetValue(msg.title)
		// p.hiddenFields["parent_target_uuid"] = struct{ val string }{val: msg.uuid}
		p.hiddenFields["parent_target_title"] = msg.title
		p.hiddenFields["parent_target_uuid"] = msg.uuid
	case data.UnauthorizedApiDataErr:
		p.cfg.logger.Error(
			msg.Error(),
			slog.Int("status", msg.Status),
			slog.String("action", "save record"),
			slog.String("type", string(p.recordType)),
		)
		p.err = errors.New(msg.Msg)
	case data.UnexpectedApiDataErr:
		p.cfg.logger.Error(
			msg.Error(),
			slog.String("action", "save record"),
			slog.String("type", string(p.recordType)),
		)
		p.err = errors.New(msg.Msg)
	case error:
		p.cfg.logger.Error(
			msg.Error(),
			slog.String("occurence", "record config page"),
		)
		p.err = msg
	}

	for i, field := range p.fields {
		retModel, retCmd := field.Update(msg)
		p.fields[i] = retModel.(Focusable)
		if retCmd != nil {
			cmds = append(cmds, retCmd)
		}
	}

	if p.selector != nil {
		var cmd tea.Cmd
		p.selector, cmd = p.selector.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return p, tea.Batch(cmds...)
}

func (p recordConfigPage) View() string {
	name := p.fields[0]
	due := p.fields[1]
	description := p.fields[2]
	status := p.fields[3]
	note := p.fields[4]
	targetSourcePrompt, targetSource := p.parentTargetField()

	titleView := style.TitleBarView([]string{p.title}, viewWidth, false)

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

	var statusPrompt string
	if status.Focused() {
		statusPrompt = fmt.Sprintf(
			"%s %s",
			style.FormFieldStyle.Prompt("Status", status.Focused()),
			style.FormFieldStyle.Helper.Render("(Use ←/→ keys to select)"),
		)
	} else {
		statusPrompt = style.FormFieldStyle.Prompt("Status", status.Focused())
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
				statusPrompt,
				style.FormFieldStyle.Content.Render(status.View()),
			),
		),

		lipgloss.NewStyle().Width(formWidth).Margin(1, 5, 0).Render(notePrompt),
	)

	var targetConfig string
	if p.recordType == data.RecordTypeAction {
		var content string
		if targetSource.View() == "" {
			content = lipgloss.NewStyle().Foreground(colors.HelperTextDim).Render("target UUID")
		} else {
			content = style.FormFieldStyle.Content.Render(targetSource.View())
		}

		targetConfig = lipgloss.NewStyle().Width(formWidth).Margin(1, 5, 0).Render(
			targetSourcePrompt + "\n" +
				// style.FormFieldStyle.Content.Render(targetSource.View()) +
				content + "\n" +
				style.FormFieldStyle.Error.Render(targetSource.Error()),
		)
		configContent = lipgloss.JoinVertical(lipgloss.Left, targetConfig, configContent)
	}

	var helperContent []style.HelperContent
	switch p.mode {
	case style.NormalView:
		helperContent = []style.HelperContent{
			{Key: "Esc", Action: "finish editing"},
			{Key: "Tab/Shift+Tab", Action: "navigate"},
		}
	case style.HighlightView:
		var saveAction string
		if p.action == cmdCreate {
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

	helperView := style.HelperView(helperContent, viewWidth, p.mode)

	msgView := style.ErrorView(style.ViewSize{Width: viewWidth, Height: 1}, p.err, nil)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		titleView,
		configContent,
		msgView,
		helperView,
	)

	if p.selector != nil {
		overlayX := lipgloss.Width(container)/2 - lipgloss.Width(p.selector.View())/2
		overlayY := lipgloss.Height(container)/2 - lipgloss.Height(p.selector.View())/2
		container = strview.PlaceOverlay(overlayX, overlayY, p.selector.View(), container)
	}

	return style.ContainerStyle(p.width, container, 5).Render(container)
}

func (p recordConfigPage) validationError() error {
	return fieldsValidation(p.fields, "input validation failed")
}

func (p recordConfigPage) create() tea.Cmd {
	title := p.fields[0].Value()
	due := p.fields[1].Value()
	description := p.fields[2].Value()
	status := p.fields[3].Value()
	note := p.fields[4].Value()

	if err := p.validationError(); err != nil {
		return validationErrorCmd(err)
	}

	d := recordRequestData{
		title:       title,
		description: description,
		status:      status,
		note:        note,
		dueDate:     due,
	}
	if p.recordType == data.RecordTypeAction {
		d.targetUUID = p.hiddenFields["parent_target_uuid"]
	}

	return p.hooks.create(p.cfg.serverURL, d, p, p.prevPage(), p.cfg.authClient)
}

func (p recordConfigPage) update() tea.Cmd {
	title := p.fields[0].Value()
	due := p.fields[1].Value()
	description := p.fields[2].Value()
	status := p.fields[3].Value()
	note := p.fields[4].Value()

	if err := p.validationError(); err != nil {
		return validationErrorCmd(err)
	}

	d := recordRequestData{
		uuid:        p.uuid,
		title:       title,
		description: description,
		status:      status,
		note:        note,
		dueDate:     due,
	}
	if p.recordType == data.RecordTypeAction {
		d.targetUUID = p.hiddenFields["parent_target_uuid"]
	}

	return p.hooks.update(p.cfg.serverURL, d, p, p.prevPage(), p.cfg.authClient)
}

func (p recordConfigPage) parentTargetField() (string, Focusable) {
	if p.recordType != data.RecordTypeAction {
		return "", nil
	}

	target := p.fields[5]

	var targetPrompt string
	if target.Focused() {
		targetPrompt = style.FormFieldStyle.Prompt("Target Source", target.Focused()) +
			" " + style.FormFieldStyle.Helper.Render("(Press 'e' to select)")
	} else {
		targetPrompt = style.FormFieldStyle.Prompt("Target Source", target.Focused())
	}

	return targetPrompt, target
}

func (p recordConfigPage) prevPage() tea.Model {
	if v, ok := p.prev.(listPage); ok {
		if v.src.uuid != "" && v.src.uuid != p.hiddenFields["parent_target_uuid"] {
			v.src.uuid = p.hiddenFields["parent_target_uuid"]
			v.src.title = p.hiddenFields["parent_target_title"]
			return v
		}
	}

	return p.prev
}
