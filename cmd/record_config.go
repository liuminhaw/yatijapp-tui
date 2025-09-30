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
		serverURL, msg string,
		body recordRequestData,
		src, redirect tea.Model,
		client *authclient.AuthClient,
	) tea.Cmd
}

type recordConfigPage struct {
	cfg    config
	action cmdAction

	uuid       string
	record     yatijappRecord
	recordType data.RecordType
	hooks      recordConfigHooks

	title        string
	fields       []Focusable
	focused      int
	focusedCache int
	hiddenFields map[string]string

	width  int
	height int

	err  error
	prev tea.Model // Previous model for navigation

	selectorFields map[data.RecordType]int
	selector       tea.Model
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
	selectorFields := make(map[data.RecordType]int)

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
	parentTarget := components.NewText(
		"",
		showSelectorMsg{selection: data.RecordTypeTarget},
	)
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

	var focusCache int
	switch recordType {
	case data.RecordTypeTarget:
		focusables[0].Focus() // Focus name field
	case data.RecordTypeAction:
		focusables = append(focusables, parentTarget)
		if record == nil {
			focusables[5].Focus() // Focus parent target field
			focusCache = 5
		} else {
			focusables[0].Focus() // Focus name field
		}
		selectorFields[data.RecordTypeTarget] = 5
	}

	return recordConfigPage{
		cfg:            cfg,
		action:         recordAction,
		focusedCache:   focusCache,
		record:         record,
		recordType:     recordType,
		uuid:           uuid,
		title:          title,
		fields:         focusables,
		hiddenFields:   hiddens,
		selectorFields: selectorFields,
		width:          size.Width,
		height:         size.Height,
		prev:           prev,
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

func newSessionConfigPage(
	cfg config,
	title string,
	size style.ViewSize,
	record yatijappRecord,
	prev tea.Model,
) (recordConfigPage, error) {
	focusables := []Focusable{}
	hiddens := make(map[string]string)
	selectorFields := make(map[data.RecordType]int)

	note := model.NewNoteModel()

	parentTarget := components.NewText("", showSelectorMsg{selection: data.RecordTypeTarget})
	parentTarget.ValidateFunc = validator.ValidateRequired("target is required")

	parentAction := components.NewText("", showSelectorMsg{selection: data.RecordTypeAction})
	parentAction.ValidateFunc = validator.ValidateRequired("action is required")

	focusables = append(focusables, parentTarget, parentAction, note)
	focusables[0].Focus() // Focus parent target field

	selectorFields[data.RecordTypeTarget] = 0
	selectorFields[data.RecordTypeAction] = 1

	if record != nil {
		parentTarget.SetValue(record.GetParentsTitle()[data.RecordTypeTarget])
		hiddens["parent_target_uuid"] = record.GetParentsUUID()[data.RecordTypeTarget]
		parentAction.SetValue(record.GetParentsTitle()[data.RecordTypeAction])
		hiddens["parent_action_uuid"] = record.GetParentsUUID()[data.RecordTypeAction]
	}

	return recordConfigPage{
		cfg:            cfg,
		action:         cmdCreate,
		record:         record,
		recordType:     data.RecordTypeSession,
		title:          title,
		fields:         focusables,
		focused:        0,
		hiddenFields:   hiddens,
		selectorFields: selectorFields,
		width:          size.Width,
		height:         size.Height,
		prev:           prev,
		hooks: recordConfigHooks{
			create: createSession,
			update: updateSession,
		},
	}, nil
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
			return p, switchToPreviousCmd(p.prev)
		case "ctrl+s":
			switch p.action {
			case cmdCreate:
				return p, p.create()
			case cmdUpdate:
				p.cfg.logger.Info("updating record", slog.String("uuid", p.uuid))
				return p, p.update()
			}
		case "ctrl+c":
			return p, tea.Quit
		}
	case showSelectorMsg:
		switch msg.selection {
		case data.RecordTypeTarget:
			p.selector = newTargetSelectorPage(p.cfg, style.ViewSize{Width: p.width, Height: p.height}, p)
		case data.RecordTypeAction:
			targetUUID := p.hiddenFields["parent_target_uuid"]
			p.selector = newActionSelectorPage(
				p.cfg, style.ViewSize{Width: p.width, Height: p.height}, targetUUID, p,
			)
		}
		cmd := p.selector.Init()
		cmds = append(cmds, cmd)
	case selectorTargetSelectedMsg:
		p.selector = nil
		if p.fields[p.focusedCache].Value() != msg.title {
			p.fields[p.focusedCache].SetValue(msg.title)
			p.hiddenFields["parent_target_title"] = msg.title
			p.hiddenFields["parent_target_uuid"] = msg.uuid
			if idx, ok := p.selectorFields[data.RecordTypeAction]; ok {
				p.fields[idx].SetValue("")
				delete(p.hiddenFields, "parent_action_title")
				delete(p.hiddenFields, "parent_action_uuid")
			}
		}
	case selectorActionSelectedMsg:
		p.selector = nil
		p.fields[p.focusedCache].SetValue(msg.title)
		p.hiddenFields["parent_action_title"] = msg.title
		p.hiddenFields["parent_action_uuid"] = msg.uuid
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
	if p.recordType == data.RecordTypeSession {
		return p.sessionCreateView()
	}

	return p.tarActConfigView()
}

func (p recordConfigPage) tarActConfigView() string {
	name := p.fields[0]
	due := p.fields[1]
	description := p.fields[2]
	status := p.fields[3]
	noteField := p.noteField(4)
	var targetField string
	if p.recordType == data.RecordTypeAction {
		targetField = p.parentField(data.RecordTypeTarget, 5, false)
	}

	titleView := style.TitleBarView([]string{p.title}, viewWidth, false)

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

		noteField+"\n",
	)

	if p.recordType == data.RecordTypeAction {
		configContent = lipgloss.JoinVertical(lipgloss.Left, targetField, configContent)
	}

	var saveAction string
	if p.action == cmdCreate {
		saveAction = "create"
	} else {
		saveAction = "save"
	}
	helperContent := []style.HelperContent{
		{Key: "Esc", Action: "back"},
		{Key: "Tab/Shift+Tab", Action: "navigate"},
		{Key: "<C-s>", Action: saveAction},
		{Key: "<C-c>", Action: "quit"},
	}

	helperView := style.HelperView(helperContent, viewWidth)

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

func (p recordConfigPage) sessionCreateView() string {
	title := lipgloss.NewStyle().
		Width(60).
		Margin(0, 2).
		AlignHorizontal(lipgloss.Center).
		Foreground(colors.Secondary).
		Bold(true).
		Render("Start new session")

	fieldStyle := lipgloss.NewStyle().Width(60).Margin(1, 3, 0)
	targetField := p.parentField(data.RecordTypeTarget, 0, true, fieldStyle)
	actionField := p.parentField(data.RecordTypeAction, 1, true, fieldStyle)
	noteField := p.noteField(2, fieldStyle)

	helper := lipgloss.NewStyle().
		Width(60).
		Margin(1, 3, 0).
		AlignHorizontal(lipgloss.Center).
		Render(lipgloss.StyleRanges(
			"Esc back    Tab/Shift+Tab navigate    <C-s> start\n",
			lipgloss.Range{Start: 0, End: 3, Style: style.Document.Primary},
			lipgloss.Range{Start: 4, End: 8, Style: style.Document.Normal},
			lipgloss.Range{Start: 12, End: 25, Style: style.Document.Primary},
			lipgloss.Range{Start: 26, End: 34, Style: style.Document.Normal},
			lipgloss.Range{Start: 38, End: 43, Style: style.Document.Primary},
			lipgloss.Range{Start: 44, End: 49, Style: style.Document.Normal},
		))

	msgView := style.ErrorView(style.ViewSize{Width: 66, Height: 1}, p.err, nil)

	form := lipgloss.JoinVertical(
		lipgloss.Left, title, targetField, actionField, noteField, msgView, helper,
	)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colors.Text).
		Render(form)
}

func (p recordConfigPage) validationError() error {
	return fieldsValidation(p.fields, "input validation failed")
}

func (p recordConfigPage) create() tea.Cmd {
	var d recordRequestData
	var cmd tea.Cmd
	switch p.recordType {
	case data.RecordTypeTarget, data.RecordTypeAction:
		d, cmd = p.tarActCreate()
	case data.RecordTypeSession:
		d, cmd = p.sessionCreate()
	default:
		panic("create record with unknown type: " + string(p.recordType))
	}

	if cmd != nil {
		return cmd
	}

	return p.hooks.create(p.cfg.serverURL, d, p, p.prevPage(), p.cfg.authClient)
}

func (p recordConfigPage) tarActCreate() (recordRequestData, tea.Cmd) {
	title := p.fields[0].Value()
	due := p.fields[1].Value()
	description := p.fields[2].Value()
	status := p.fields[3].Value()
	note := p.fields[4].Value()

	if err := p.validationError(); err != nil {
		return recordRequestData{}, validationErrorCmd(err)
	}

	d := recordRequestData{
		title:       title,
		description: description,
		status:      status,
		note:        note,
		dueDate:     due,
	}
	if p.recordType == data.RecordTypeAction {
		targetUUID := p.hiddenFields["parent_target_uuid"]
		d.targetUUID = targetUUID
	}

	return d, nil
}

func (p recordConfigPage) sessionCreate() (recordRequestData, tea.Cmd) {
	targetUUID := p.hiddenFields["parent_target_uuid"]
	actionUUID := p.hiddenFields["parent_action_uuid"]
	note := p.fields[2].Value()

	if err := p.validationError(); err != nil {
		return recordRequestData{}, validationErrorCmd(err)
	}

	if targetUUID == "" {
		return recordRequestData{}, validationErrorCmd(errors.New("target is required"))
	}
	if actionUUID == "" {
		return recordRequestData{}, validationErrorCmd(errors.New("action is required"))
	}

	d := recordRequestData{
		targetUUID: targetUUID,
		actionUUID: actionUUID,
		note:       note,
	}

	return d, nil
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

	return p.hooks.update(p.cfg.serverURL, "", d, p, p.prevPage(), p.cfg.authClient)
}

func (p recordConfigPage) noteField(fieldIndex int, styling ...lipgloss.Style) string {
	if len(p.fields) <= fieldIndex {
		return ""
	}

	note := p.fields[fieldIndex]
	var notePrompt string
	if note.Focused() {
		notePrompt = fmt.Sprintf(
			"%s %s\n%s",
			style.FormFieldStyle.Prompt("Note", note.Focused()),
			style.FormFieldStyle.Helper.Render("(Press 'e' to edit)"),
			style.FormFieldStyle.Error.Render(note.Error()),
		)
	} else {
		notePrompt = fmt.Sprintf(
			"%s\n%s",
			style.FormFieldStyle.Prompt("Note", note.Focused()),
			style.FormFieldStyle.Error.Render(note.Error()),
		)
	}

	applyStyling := lipgloss.NewStyle().Width(formWidth).Margin(1, 5, 0)
	for _, s := range styling {
		applyStyling = s.Inherit(applyStyling)
	}

	return applyStyling.Render(notePrompt)
}

func (p recordConfigPage) parentField(
	parentType data.RecordType,
	index int,
	compacted bool,
	styling ...lipgloss.Style,
) string {
	if len(p.fields) <= index {
		return ""
	}

	focusable := p.fields[index]
	placeholder := string(parentType)

	var prompt string
	if focusable.Focused() {
		prompt = style.FormFieldStyle.Prompt(placeholder+" Source", true) +
			" " + style.FormFieldStyle.Helper.Render("(Press 'e' to select)")
	} else {
		prompt = style.FormFieldStyle.Prompt(placeholder+" Source", false)
	}

	var input string
	if focusable.View() == "" {
		input = lipgloss.NewStyle().Foreground(colors.HelperTextDim).Render(placeholder + " name")
	} else {
		input = style.FormFieldStyle.Content.Render(focusable.View())
	}

	applyStyling := lipgloss.NewStyle().Width(formWidth).Margin(1, 5, 0)

	for _, s := range styling {
		applyStyling = s.Inherit(applyStyling)
	}

	if compacted {
		return applyStyling.Render(
			prompt + " " + style.FormFieldStyle.Error.Render(focusable.Error()) +
				"\n" + input,
		)
	} else {
		return applyStyling.Render(
			prompt + "\n" + input + "\n" + style.FormFieldStyle.Error.Render(focusable.Error()),
		)
	}
}

func (p recordConfigPage) prevPage() tea.Model {
	if v, ok := p.prev.(listPage); ok {
		parentType := v.recordType.GetParentType()
		if parentType == "" {
			return p.prev
		}

		switch v.recordType {
		case data.RecordTypeAction:
			targetUUID := v.src.UUID(data.RecordTypeTarget)
			if targetUUID != "" && targetUUID != p.hiddenFields["parent_target_uuid"] {
				v.src[data.RecordTypeTarget] = data.RecordParent{
					UUID:  p.hiddenFields["parent_target_uuid"],
					Title: p.hiddenFields["parent_target_title"],
				}
			}
		case data.RecordTypeSession:
			targetUUID := v.src.UUID(data.RecordTypeTarget)
			if targetUUID != "" && targetUUID != p.hiddenFields["parent_target_uuid"] {
				v.src[data.RecordTypeTarget] = data.RecordParent{
					UUID:  p.hiddenFields["parent_target_uuid"],
					Title: p.hiddenFields["parent_target_title"],
				}
			}

			actionUUID := v.src.UUID(data.RecordTypeAction)
			if actionUUID != "" && actionUUID != p.hiddenFields["parent_action_uuid"] {
				v.src[data.RecordTypeAction] = data.RecordParent{
					UUID:  p.hiddenFields["parent_action_uuid"],
					Title: p.hiddenFields["parent_action_title"],
				}
			}
		}
	}

	return p.prev
}
