package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/model"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
	"github.com/liuminhaw/yatijapp-tui/pkg/strview"
)

type listHooks struct {
	loadAll func(info data.ListRequestInfo, msg, src string, client *authclient.AuthClient) tea.Cmd
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

	selection  recordsSelection
	hooks      listHooks
	recordType data.RecordType

	width  int
	height int

	spinner spinner.Model
	loading bool

	msg   string
	error error

	src  data.RecordParents
	prev tea.Model // Previous model for navigation

	popupModels []tea.Model
	popup       string
}

func newListPage(cfg config, termSize style.ViewSize, prev tea.Model) listPage {
	return listPage{
		cfg:         cfg,
		selection:   newRecordsSelection(10),
		width:       termSize.Width,
		height:      termSize.Height,
		spinner:     spinner.New(spinner.WithSpinner(spinner.Line)),
		loading:     true,
		prev:        prev,
		popupModels: []tea.Model{},
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

func (l listPage) Init() tea.Cmd {
	return tea.Batch(
		l.spinner.Tick,
		l.hooks.loadAll(
			data.ListRequestInfo{
				ServerURL:    l.cfg.apiEndpoint,
				SrcUUID:      l.src[l.recordType.GetParentType()].UUID,
				QueryStrings: l.selection.query,
			},
			"", "list", l.cfg.authClient,
		),
	)
}

func (l listPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	var popupModel tea.Model
	switch msg := msg.(type) {
	case showSelectorMsg:
		switch msg.selection {
		case data.RecordTypeTarget:
			popupModel = newTargetSelectorPage(l.cfg, style.ViewSize{Width: l.width, Height: l.height}, l)
		case data.RecordTypeAction:
			targetUUID := ""
			if configModel, ok := l.popupModels[len(l.popupModels)-1].(recordConfigPage); ok {
				targetUUID = configModel.hiddenFields["parent_target_uuid"]
			} else {
				panic("action selector without target config model as popupModel")
			}
			popupModel = newActionSelectorPage(l.cfg, style.ViewSize{Width: l.width, Height: l.height}, targetUUID, l)
		}
		l.popupModels = append(l.popupModels, popupModel)
		cmd := popupModel.Init()
		cmds = append(cmds, cmd)
		l.popup = l.popupModels[len(l.popupModels)-1].View()
	case tea.KeyMsg:
		switch msg.String() {
		case "?":
			l.clearMsg()
			l.helper = !l.helper
			if l.helper {
				l.helperPopup(20)
			} else {
				l.popup = ""
			}
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height
	case tea.KeyMsg:
		if l.popup != "" {
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return l, tea.Quit
		case "up", "k":
			l.clearMsg()
			return l, l.selection.prev()
		case "down", "j":
			l.clearMsg()
			return l, l.selection.next()
		case "right", "l":
			l.clearMsg()
			return l, l.selection.nextPage()
		case "left", "h":
			l.clearMsg()
			return l, l.selection.prevPage()
		case "enter":
			if len(l.selection.records) > 0 {
				selected := l.selection.current()
				if selected == nil {
					l.error = errors.New("no session selected on enter key press")
					return l, nil
				}

				if l.recordType == data.RecordTypeSession {
					session := selected.(data.Session)
					if session.EndsAt.Valid {
						return l, nil
					}

					updateCmd := l.hooks.update(
						l.cfg.apiEndpoint,
						"Session ended",
						recordRequestData{
							uuid:       selected.GetUUID(),
							endsAt:     sql.NullTime{Valid: true, Time: time.Now()},
							actionUUID: selected.GetParentsUUID()[data.RecordTypeAction],
						},
						l, l,
						l.cfg.authClient,
					)
					popupModel = model.NewAlert(
						"Confirm End Session",
						"confirmation",
						[]string{"Proceed to end session \"" + selected.GetTitle() + "\"?"},
						[]string{""},
						60,
						map[string]tea.Cmd{"confirm": updateCmd, "cancel": cancelPopupCmd},
					)
					l.popupModels = append(l.popupModels, popupModel)
					l.popup = l.popupModels[len(l.popupModels)-1].View()
					return l, confirmationCmd
				}
				return l, switchToRecordsCmd(selected)
			}
		case "<":
			if l.error != nil {
				l.error = nil
				return l, nil
			}
			return l, switchToPreviousCmd(l.prev)
		case "v":
			if l.selection.hasRecords() {
				return l, switchToViewCmd(l.recordType, l.selection.current().GetUUID())
			}
		case "e":
			if l.selection.hasRecords() {
				selected := l.selection.current()
				if selected == nil {
					l.error = errors.New("no session selected on enter key press")
					return l, nil
				}

				l.loading = true
				l.clearMsg()
				return l, l.hooks.load(l.cfg.apiEndpoint, selected.GetUUID(), "", l.cfg.authClient)
			}
		case "d":
			selected := l.selection.current()
			if selected == nil {
				return l, nil
			}

			var prompts, warnings []string
			switch l.recordType {
			case data.RecordTypeTarget:
				prompts = []string{"Proceed to delete target \"" + selected.GetTitle() + "\"?"}
				warnings = []string{"All actions and sessions under this target will be deleted as well."}
			case data.RecordTypeAction:
				prompts = []string{"Proceed to delete action \"" + selected.GetTitle() + "\"?"}
				warnings = []string{"All sessions under this action will be deleted as well."}
			case data.RecordTypeSession:
				if selected.GetStatus() == "completed" {
					prompts = []string{
						"Proceed to delete session",
						"\"" + selected.GetTitle() + "\"?",
					}
				} else {
					prompts = []string{"Proceed to delete session \"" + selected.GetTitle() + "\"?"}
				}
			}
			deleteCmd := l.hooks.delete(l.cfg.apiEndpoint, selected.GetUUID(), l.cfg.authClient)
			popupModel = model.NewAlert(
				"Confirm Deletion", "confirmation", prompts, warnings, 60,
				map[string]tea.Cmd{"confirm": deleteCmd, "cancel": cancelPopupCmd},
			)
			l.popupModels = append(l.popupModels, popupModel)
			l.popup = l.popupModels[len(l.popupModels)-1].View()
			return l, confirmationCmd
		case "n":
			return l, switchToCreateCmd(l.recordType, l.src)
		case "m":
			return l, switchToMenuCmd
		case "ctrl+r":
			l.clearMsg()
			return l, l.hooks.loadAll(
				data.ListRequestInfo{
					ServerURL:    l.cfg.apiEndpoint,
					SrcUUID:      l.src[l.recordType.GetParentType()].UUID,
					QueryStrings: l.selection.query,
				},
				"Records refreshed", "list", l.cfg.authClient,
			)
		}
	case cancelPopupMsg:
		l.popupModels = l.popupModels[:len(l.popupModels)-1]
		if len(l.popupModels) == 0 {
			l.popup = ""
		} else {
			l.popup = l.popupModels[len(l.popupModels)-1].View()
		}
		return l, nil
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
		popupModel, err = newSessionConfigPage(
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
		l.cfg.logger.Info("new session popup", slog.Any("record", l.popupModels))

		l.popupModels = append(l.popupModels, popupModel)
		l.popup = l.popupModels[len(l.popupModels)-1].View()
	case switchToPreviousMsg:
		l.loading = true
		l.clearMsg()
		return l, tea.Batch(
			l.spinner.Tick,
			l.hooks.loadAll(
				data.ListRequestInfo{
					ServerURL:    l.cfg.apiEndpoint,
					SrcUUID:      l.src.UUID(l.recordType.GetParentType()),
					QueryStrings: l.selection.query,
				},
				"", "list", l.cfg.authClient,
			),
		)
	case allRecordsLoadedMsg:
		// l.cfg.logger.Info("all records loaded msg received", slog.String("src", msg.src))
		if msg.src != "list" {
			break
		}
		l.msg = msg.msg
		l.selection.setRecords(msg, l.cfg.logger)
		l.loading = false
	case getRecordLoadedMsg:
		l.loading = false
		return l, switchToEditCmd(l.recordType, msg.record)
	case recordDeletedMsg:
		l.popupModels = l.popupModels[:len(l.popupModels)-1]
		if len(l.popupModels) == 0 {
			l.popup = ""
		} else {
			l.popup = l.popupModels[len(l.popupModels)-1].View()
		}
		l.clearMsg()
		return l, l.hooks.loadAll(
			data.ListRequestInfo{
				ServerURL:    l.cfg.apiEndpoint,
				SrcUUID:      l.src.UUID(l.recordType.GetParentType()),
				QueryStrings: l.selection.query,
			},
			string(msg), "list", l.cfg.authClient,
		)
	case confirmationMsg:
		l.loading = false
	case apiSuccessResponseMsg:
		l.loading = true
		l.clearMsg()
		l.cfg.logger.Info("api success", slog.Any("src", l.src))
		return l, l.hooks.loadAll(
			data.ListRequestInfo{
				ServerURL:    l.cfg.apiEndpoint,
				SrcUUID:      l.src.UUID(l.recordType.GetParentType()),
				QueryStrings: l.selection.query,
			},
			msg.msg, "list", l.cfg.authClient,
		)
	case loadMoreRecordsMsg:
		// l.cfg.logger.Info("load more records for list page")
		l.loading = true
		l.clearMsg()
		return l, l.hooks.loadAll(
			data.ListRequestInfo{
				ServerURL:    l.cfg.apiEndpoint,
				SrcUUID:      l.src.UUID(l.recordType.GetParentType()),
				QueryStrings: l.selection.query,
				Events:       []string{msg.direction},
			},
			"", "list", l.cfg.authClient,
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

	l.selection.p, cmd = l.selection.p.Update(msg)
	cmds = append(cmds, cmd)
	if len(l.popupModels) > 0 {
		lastIndex := len(l.popupModels) - 1
		l.popupModels[lastIndex], cmd = l.popupModels[lastIndex].Update(msg)
		l.popup = l.popupModels[lastIndex].View()
		cmds = append(cmds, cmd)
	}

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
	start, end := l.selection.p.GetSliceBounds(len(l.selection.records))
	for i, record := range l.selection.records[start:end] {
		content.WriteString(
			record.ListItemView(
				l.src[l.recordType.GetParentType()] == data.RecordParent{},
				i+start == l.selection.selected,
				viewWidth,
			),
		)
	}

	for i := len(l.selection.records[start:end]); i < l.selection.p.PerPage; i++ {
		content.WriteString("\n")
	}

	contentView := lipgloss.NewStyle().Render(content.String())

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		contentView,
		l.selection.view(),
	)

	if len(l.selection.records) > 0 {
		selected := l.selection.current()
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
		"<":     "Back",
		"↑/↓":   "Navigate",
		"q":     "Quit",
		"n":     "New " + strings.ToLower(recordType),
		"v":     "View " + strings.ToLower(recordType),
		"e":     "Edit " + strings.ToLower(recordType),
		"d":     "Delete " + strings.ToLower(recordType),
		"f":     "Filter",
		"m":     "Menu",
		"<C-r>": "Refresh",
		"?":     "Toggle helper",
	}
	order := []string{"<", "↑/↓", "q", "n", "v", "e", "d", "f", "m", "<C-r>", "?"}

	var enterValue string
	if l.recordType == data.RecordTypeSession && l.selection.hasRecords() &&
		!l.selection.current().(data.Session).EndsAt.Valid {
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
		{Key: "<", Action: "back"},
		{Key: "↑/↓", Action: "navigate"},
		{Key: "q", Action: "quit"},
		{Key: "?", Action: "toggle helper"},
	}

	enterValue := ""
	if l.recordType == data.RecordTypeSession && l.selection.hasRecords() &&
		!l.selection.current().(data.Session).EndsAt.Valid {
		enterValue = "end session"
	} else if l.recordType != data.RecordTypeSession {
		enterValue = "select"
	}

	if enterValue != "" {
		content = slices.Insert(content, 2, style.HelperContent{Key: "Enter", Action: enterValue})
	}

	return style.HelperView(content, width)
}
