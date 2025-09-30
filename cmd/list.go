package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
	"github.com/liuminhaw/yatijapp-tui/pkg/strview"
)

type listHooks struct {
	loadAll func(serverURL, srcUUID, msg string, client *authclient.AuthClient) tea.Cmd
	load    func(serverURL, uuid, msg string, client *authclient.AuthClient) tea.Cmd
	delete  func(serverURL, uuid string, client *authclient.AuthClient) tea.Cmd
	update  func(
		serverURL, msg string,
		d recordRequestData,
		src, redirect tea.Model,
		client *authclient.AuthClient,
	) tea.Cmd
}

type listPage struct {
	cfg    config
	helper bool

	records    []yatijappRecord
	paginator  paginator.Model
	selected   int
	hooks      listHooks
	recordType data.RecordType

	width  int
	height int

	spinner      spinner.Model
	loading      bool
	confirmation *style.ConfirmCheck

	msg   string
	error error

	src  data.RecordParents
	prev tea.Model // Previous model for navigation

	popupModel tea.Model
	popup      string
}

func newListPage(cfg config, termSize style.ViewSize, prev tea.Model) listPage {
	return listPage{
		cfg:       cfg,
		paginator: style.NewPaginator(10),
		width:     termSize.Width,
		height:    termSize.Height,
		spinner:   spinner.New(spinner.WithSpinner(spinner.Line)),
		loading:   true,
		prev:      prev,
	}
}

func newTargetListPage(
	cfg config,
	termSize style.ViewSize,
	src data.RecordParents,
	prev tea.Model,
) listPage {
	page := newListPage(cfg, termSize, prev)
	page.recordType = data.RecordTypeTarget
	page.src = src
	page.hooks = listHooks{
		loadAll: loadAllTargets,
		load:    loadTarget,
		delete:  deleteTarget,
	}

	return page
}

func newActionListPage(
	cfg config,
	termSize style.ViewSize,
	src data.RecordParents,
	prev tea.Model,
) listPage {
	cfg.logger.Info("Creating new action list page", slog.Any("src", src))

	page := newListPage(cfg, termSize, prev)
	page.recordType = data.RecordTypeAction
	page.src = src
	page.hooks = listHooks{
		loadAll: loadAllActions,
		load:    loadAction,
		delete:  deleteAction,
	}

	return page
}

func newSessionListPage(
	cfg config,
	termSize style.ViewSize,
	src data.RecordParents,
	prev tea.Model,
) listPage {
	page := newListPage(cfg, termSize, prev)
	page.recordType = data.RecordTypeSession
	page.src = src
	page.hooks = listHooks{
		loadAll: loadAllSessions,
		load:    loadSession,
		delete:  deleteSession,
		update:  updateSession,
	}

	return page
}

func (l *listPage) setConfirmation(prompt, warning string, cmd tea.Cmd) {
	l.confirmation = &style.ConfirmCheck{Prompt: prompt, Warning: warning, Cmd: cmd}
}

func (l listPage) Init() tea.Cmd {
	return tea.Batch(
		l.spinner.Tick,
		l.hooks.loadAll(
			l.cfg.serverURL,
			l.src[l.recordType.GetParentType()].UUID,
			"",
			l.cfg.authClient,
		),
	)
}

func (l listPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case showSelectorMsg:
		switch msg.selection {
		case data.RecordTypeTarget:
			l.popupModel = newTargetSelectorPage(l.cfg, style.ViewSize{Width: l.width, Height: l.height}, l)
		case data.RecordTypeAction:
			targetUUID := ""
			if configModel, ok := l.popupModel.(recordConfigPage); ok {
				targetUUID = configModel.hiddenFields["parent_target_uuid"]
			} else {
				panic("action selector without target config model as popupModel")
			}
			l.popupModel = newActionSelectorPage(
				l.cfg, style.ViewSize{Width: l.width, Height: l.height}, targetUUID, l,
			)
		}
		cmd := l.popupModel.Init()
		cmds = append(cmds, cmd)
		l.popup = l.popupModel.View()
	}
	if l.popupModel != nil {
		var cmd tea.Cmd
		l.popupModel, cmd = l.popupModel.Update(msg)
		l.popup = l.popupModel.View()
		cmds = append(cmds, cmd)

		return l, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height
	case tea.KeyMsg:
		if l.confirmation != nil {
			switch msg.String() {
			case "y":
				execCmd := l.confirmation.Cmd
				l.confirmation = nil
				l.popup = ""
				l.clearMsg()
				return l, execCmd
			case "n":
				l.confirmation = nil
				l.popup = ""
				return l, nil
			default:
				return l, nil
			}
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return l, tea.Quit
		case "up", "k":
			l.clearMsg()
			if l.selected > 0 {
				l.selected--
				start, _ := l.paginator.GetSliceBounds(len(l.records))
				if l.selected < start {
					l.paginator.PrevPage()
				}
			}
		case "down", "j":
			l.clearMsg()
			if l.selected < len(l.records)-1 {
				l.selected++
				_, end := l.paginator.GetSliceBounds(len(l.records))
				if l.selected >= end {
					l.paginator.NextPage()
				}
			}
		case "right", "l":
			l.clearMsg()
			if l.paginator.OnLastPage() {
				break
			}
			l.paginator.NextPage()
			l.selected, _ = l.paginator.GetSliceBounds(len(l.records))
		case "left", "h":
			l.clearMsg()
			if l.paginator.OnFirstPage() {
				break
			}
			l.paginator.PrevPage()
			l.selected, _ = l.paginator.GetSliceBounds(len(l.records))
		case "enter":
			if len(l.records) > 0 {
				if l.recordType == data.RecordTypeSession {
					session := l.records[l.selected].(data.Session)
					if session.EndsAt.Valid {
						return l, nil
					}

					updateCmd := l.hooks.update(
						l.cfg.serverURL,
						"Session ended",
						recordRequestData{
							uuid:       l.records[l.selected].GetUUID(),
							endsAt:     sql.NullTime{Valid: true, Time: time.Now()},
							actionUUID: l.records[l.selected].GetParentsUUID()[data.RecordTypeAction],
						},
						l, l,
						l.cfg.authClient,
					)
					l.setConfirmation("Proceed to end session \""+l.records[l.selected].GetTitle()+"\"?", "", updateCmd)
					l.popup = l.confirmation.View("Confirm End Session", 60)
					return l, confirmationCmd
				}
				return l, switchToRecordsCmd(l.records[l.selected])
			}
		case "<":
			if l.error != nil {
				l.error = nil
				return l, nil
			}
			return l, switchToPreviousCmd(l.prev)
		case "?":
			l.clearMsg()
			l.helper = !l.helper
			if l.helper {
				l.helperPopup(20)
			} else {
				l.popup = ""
			}
		case "v":
			if len(l.records) > 0 {
				return l, switchToViewCmd(l.recordType, l.records[l.selected].GetUUID())
			}
		case "e":
			if len(l.records) > 0 {
				l.loading = true
				l.clearMsg()
				return l, l.hooks.load(l.cfg.serverURL, l.records[l.selected].GetUUID(), "", l.cfg.authClient)
			}
		case "d":
			if len(l.records) == 0 {
				return l, nil
			}

			var prompt, warning string
			switch l.recordType {
			case data.RecordTypeTarget:
				prompt = "Proceed to delete target \"" + l.records[l.selected].GetTitle() + "\"?"
				warning = "All actions and sessions under this target will be deleted as well."
			case data.RecordTypeAction:
				prompt = "Proceed to delete action \"" + l.records[l.selected].GetTitle() + "\"?"
				warning = "All sessions under this action will be deleted as well."
			case data.RecordTypeSession:
				prompt = "Proceed to delete session \"" + l.records[l.selected].GetTitle() + "\"?"
			}
			deleteCmd := l.hooks.delete(l.cfg.serverURL, l.records[l.selected].GetUUID(), l.cfg.authClient)
			l.setConfirmation(prompt, warning, deleteCmd)
			l.popup = l.confirmation.View("Confirm Deletion", 60)
			return l, confirmationCmd
		case "n":
			return l, switchToCreateCmd(l.recordType, l.src)
		case "ctrl+r":
			l.clearMsg()
			return l, l.hooks.loadAll(
				l.cfg.serverURL,
				l.src[l.recordType.GetParentType()].UUID,
				"Records refreshed",
				l.cfg.authClient,
			)
		}
	case showSessionCreateMsg:
		var record yatijappRecord
		if !msg.parents.IsEmpty() {
			record = data.Session{
				TargetUUID:  msg.parents.UUID(data.RecordTypeTarget),
				TargetTitle: msg.parents.Title(data.RecordTypeTarget),
				ActionUUID:  msg.parents.UUID(data.RecordTypeAction),
				ActionTitle: msg.parents.Title(data.RecordTypeAction),
			}
		}

		l.cfg.logger.Info("new session config page", slog.Any("src", l.src))
		var err error
		l.popupModel, err = newSessionConfigPage(
			l.cfg,
			"New Session",
			style.ViewSize{Width: l.width, Height: l.height},
			record,
			l,
		)
		if err != nil {
			l.cfg.logger.Error(err.Error(), slog.String("action", "show new session popup"))
			return l, tea.Quit
		}
		l.cfg.logger.Info("new session popup", slog.Any("record", l.popupModel))

		l.popup = l.popupModel.View()
	case switchToPreviousMsg:
		l.loading = true
		l.clearMsg()
		return l, l.hooks.loadAll(
			l.cfg.serverURL,
			l.src.UUID(l.recordType.GetParentType()),
			"",
			l.cfg.authClient,
		)
	case allRecordsLoadedMsg:
		l.msg = msg.msg
		l.records = msg.records
		l.paginator.SetTotalPages(len(l.records))
		if l.paginator.Page > l.paginator.TotalPages-1 {
			l.paginator.Page = l.paginator.TotalPages - 1
		}
		_, end := l.paginator.GetSliceBounds(len(l.records))
		if l.selected >= end && l.selected > 0 {
			l.selected = end - 1
		}
		l.loading = false
	case getRecordLoadedMsg:
		l.loading = false
		return l, switchToEditCmd(l.recordType, msg.record)
	case recordDeletedMsg:
		l.clearMsg()
		return l, l.hooks.loadAll(
			l.cfg.serverURL,
			l.src.UUID(l.recordType.GetParentType()),
			string(msg),
			l.cfg.authClient,
		)
	case confirmationMsg:
		l.loading = false
	case apiSuccessResponseMsg:
		l.loading = true
		l.clearMsg()
		l.cfg.logger.Info("api success", slog.Any("src", l.src))
		return l, l.hooks.loadAll(
			l.cfg.serverURL, l.src.UUID(l.recordType.GetParentType()), msg.msg, l.cfg.authClient,
		)
	case data.UnauthorizedApiDataErr:
		l.cfg.logger.Error(
			msg.Error(),
			slog.Int("status", msg.Status),
			slog.String("occurence", "list page"),
			slog.String("type", string(l.recordType)),
		)
		l.loading = false
		return l, switchToMenuCmd
	case data.UnexpectedApiDataErr:
		l.cfg.logger.Error(
			msg.Error(),
			slog.String("occurence", "list page"),
			slog.String("type", string(l.recordType)),
		)
		l.error = errors.New(msg.Msg)
		l.loading = false
	case error:
		l.cfg.logger.Error(msg.Error(), slog.String("occurence", "list page"))
		l.error = msg
		l.loading = false
	case spinner.TickMsg:
		l.spinner, cmd = l.spinner.Update(msg)
		return l, cmd
	}

	l.paginator, cmd = l.paginator.Update(msg)
	cmds = append(cmds, cmd)

	return l, tea.Batch(cmds...)
}

func (l listPage) View() string {
	if l.loading {
		container := style.LoadingView(
			&l.spinner,
			"Loading list",
			style.ViewSize{Width: viewWidth, Height: 10},
		)

		return style.ContainerStyle(l.width, container, 5).Render(container)
	}

	var title string
	if l.msg == "" {
		switch l.recordType {
		case data.RecordTypeTarget:
			title = style.TitleBarView([]string{"Targets"}, viewWidth, false)
		case data.RecordTypeAction:
			if t := l.src.Title(l.recordType.GetParentType()); t != "" {
				title = style.TitleBarView([]string{t, "Actions"}, viewWidth, false)
			} else {
				title = style.TitleBarView([]string{"Actions"}, viewWidth, false)
			}
		case data.RecordTypeSession:
			if a := l.src.Title(l.recordType.GetParentType()); a != "" {
				t := l.src.Title(data.RecordTypeTarget)
				title = style.TitleBarView([]string{t, a, "Sessions"}, viewWidth, false)
			} else {
				title = style.TitleBarView([]string{"Sessions"}, viewWidth, false)
			}
		default:
			panic("unsupported record type in list page view")
		}
	} else {
		title = style.TitleBarView([]string{l.msg}, viewWidth, true)
	}

	if l.error != nil {
		return style.FullPageErrorView(
			title,
			l.width,
			style.ViewSize{Width: viewWidth, Height: 16},
			l.error,
			[]style.HelperContent{{Key: "<", Action: "back"}},
		)
	}

	helperView := l.listPageHelper(viewWidth)

	var content strings.Builder
	start, end := l.paginator.GetSliceBounds(len(l.records))
	for i, record := range l.records[start:end] {
		content.WriteString(
			record.ListItemView(
				l.src[l.recordType.GetParentType()] == data.RecordParent{},
				i+start == l.selected,
				viewWidth,
			),
		)
	}

	for i := len(l.records[start:end]); i < l.paginator.PerPage; i++ {
		content.WriteString("\n")
	}

	contentView := lipgloss.NewStyle().Render(content.String())

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		contentView,
		l.paginator.View(),
	)

	if len(l.records) > 0 {
		selected := l.records[l.selected]
		detailView := data.ListPageDetailView(
			selected.ListItemDetailView(
				l.src[l.recordType.GetParentType()] == data.RecordParent{},
				viewWidth,
			),
			l.popup != "",
		)
		container = lipgloss.JoinVertical(lipgloss.Center, container, detailView)
	}
	container = lipgloss.JoinVertical(lipgloss.Center, container, helperView)

	if l.popup != "" {
		overlayX := lipgloss.Width(container)/2 - lipgloss.Width(l.popup)/2
		overlayY := lipgloss.Height(container)/2 - lipgloss.Height(l.popup)/2
		container = strview.PlaceOverlay(overlayX, overlayY, l.popup, container)
	}
	return style.ContainerStyle(l.width, container, 5).Render(container)
}

func (l *listPage) clearMsg() {
	l.msg = ""
}

func (l *listPage) helperPopup(width int) {
	recordType := string(l.recordType)
	items := map[string]string{
		"<":     "Back to menu",
		"↑/↓":   "Navigate",
		"q":     "Quit",
		"n":     "New " + strings.ToLower(recordType),
		"v":     "View " + strings.ToLower(recordType),
		"e":     "Edit " + strings.ToLower(recordType),
		"d":     "Delete " + strings.ToLower(recordType),
		"f":     "Filter",
		"<C-r>": "Refresh",
		"?":     "Toggle helper",
	}
	order := []string{"<", "↑/↓", "q", "n", "v", "e", "d", "f", "<C-r>", "?"}

	var enterValue string
	if l.recordType == data.RecordTypeSession && len(l.records) > 0 &&
		!l.records[l.selected].(data.Session).EndsAt.Valid {
		enterValue = "End session"
	} else if l.recordType != data.RecordTypeSession {
		enterValue = "Select"
	}

	if enterValue != "" {
		items["Enter"] = enterValue
		order = slices.Insert(order, 2, "Enter")
	}

	l.popup = style.FullHelpView([]style.FullHelpContent{
		{
			Title:        "Key Maps",
			Items:        items,
			Order:        order,
			KeyHighlight: true,
		},
	}, width)
}

func (l listPage) listPageHelper(width int) string {
	content := []style.HelperContent{
		{Key: "<", Action: "menu"},
		{Key: "↑/↓", Action: "navigate"},
		{Key: "q", Action: "quit"},
		{Key: "?", Action: "toggle helper"},
	}

	enterValue := ""
	if l.recordType == data.RecordTypeSession && len(l.records) > 0 &&
		!l.records[l.selected].(data.Session).EndsAt.Valid {
		enterValue = "end session"
	} else if l.recordType != data.RecordTypeSession {
		enterValue = "select"
	}

	if enterValue != "" {
		content = slices.Insert(content, 2, style.HelperContent{Key: "Enter", Action: enterValue})
	}

	return style.HelperView(content, width)
}
