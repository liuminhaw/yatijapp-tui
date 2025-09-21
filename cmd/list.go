package main

import (
	"errors"
	"log/slog"
	"strings"

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

	// srcUUID string
	src  sourceInfo
	prev tea.Model // Previous model for navigation
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
	src sourceInfo,
	// srcUUID string,
	prev tea.Model,
) listPage {
	page := newListPage(cfg, termSize, prev)
	page.recordType = data.RecordTypeTarget
	// page.srcUUID = srcUUID
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
	// srcUUID string,
	src sourceInfo,
	prev tea.Model,
) listPage {
	page := newListPage(cfg, termSize, prev)
	page.recordType = data.RecordTypeAction
	// page.srcUUID = srcUUID
	page.src = src
	page.hooks = listHooks{
		loadAll: loadAllActions,
		load:    loadAction,
		delete:  deleteAction,
	}

	return page
}

func (l *listPage) setConfirmation(prompt, warning string) {
	l.confirmation = &style.ConfirmCheck{Prompt: prompt, Warning: warning}
}

func (l listPage) Init() tea.Cmd {
	return tea.Batch(
		l.spinner.Tick,
		l.hooks.loadAll(l.cfg.serverURL, l.src.uuid, "", l.cfg.authClient),
	)
}

func (l listPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height
	case tea.KeyMsg:
		if l.confirmation != nil {
			switch msg.String() {
			case "y":
				l.confirmation = nil
				l.clearMsg()
				return l, l.hooks.delete(l.cfg.serverURL, l.records[l.selected].GetUUID(), l.cfg.authClient)
			case "n":
				l.confirmation = nil
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
				return l, switchToRecordsCmd(
					l.recordType,
					l.records[l.selected].GetUUID(),
					l.records[l.selected].GetTitle(),
				)
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
			}
			l.setConfirmation(prompt, warning)
			return l, confirmationCmd
		case "n":
			return l, switchToCreateCmd(l.recordType, l.src.uuid, l.src.title)
		case "ctrl+r":
			l.clearMsg()
			return l, l.hooks.loadAll(l.cfg.serverURL, l.src.uuid, "Records refreshed", l.cfg.authClient)
		}
	case switchToPreviousMsg:
		l.loading = true
		l.clearMsg()
		return l, l.hooks.loadAll(l.cfg.serverURL, l.src.uuid, "", l.cfg.authClient)
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
		return l, l.hooks.loadAll(l.cfg.serverURL, l.src.uuid, string(msg), l.cfg.authClient)
	case confirmationMsg:
		l.loading = false
	case apiSuccessResponseMsg:
		l.loading = true
		l.clearMsg()
		return l, l.hooks.loadAll(l.cfg.serverURL, l.src.uuid, msg.msg, l.cfg.authClient)
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
	return l, cmd
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
			if l.src != (sourceInfo{}) {
				title = style.TitleBarView([]string{l.src.title, "Actions"}, viewWidth, false)
			} else {
				title = style.TitleBarView([]string{"Actions"}, viewWidth, false)
			}
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

	var popup string
	if l.confirmation != nil {
		popup = l.confirmation.View("Confirm Deletion", 60)
	} else if l.helper {
		popup = helperPopup(20)
	}

	helperView := listPageHelper(viewWidth)

	var content strings.Builder
	start, end := l.paginator.GetSliceBounds(len(l.records))
	for i, record := range l.records[start:end] {
		content.WriteString(
			record.ListItemView(l.src == sourceInfo{}, i+start == l.selected, viewWidth),
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
			selected.ListItemDetailView(l.src == sourceInfo{}, viewWidth),
			popup != "",
		)
		container = lipgloss.JoinVertical(lipgloss.Center, container, detailView)
	}
	container = lipgloss.JoinVertical(lipgloss.Center, container, helperView)

	if popup != "" {
		overlayX := lipgloss.Width(container)/2 - lipgloss.Width(popup)/2
		overlayY := lipgloss.Height(container)/2 - lipgloss.Height(popup)/2
		container = strview.PlaceOverlay(overlayX, overlayY, popup, container)
	}
	return style.ContainerStyle(l.width, container, 5).Render(container)
}

func (l *listPage) clearMsg() {
	l.msg = ""
}

func helperPopup(width int) string {
	return style.FullHelpView([]style.FullHelpContent{
		{
			Title: "Key Maps",
			Items: map[string]string{
				"<":     "Back to menu",
				"↑/↓":   "Navigate",
				"Enter": "Select",
				"q":     "Quit",
				"n":     "New target",
				"v":     "View target",
				"e":     "Edit target",
				"d":     "Delete target",
				"f":     "Filter",
				"<C-r>": "Refresh",
				"?":     "Toggle helper",
			},
			Order: []string{
				"<", "↑/↓", "Enter", "q", "n", "v", "e", "d", "f", "<C-r>", "?",
			},
			KeyHighlight: true,
		},
	}, width)
}

func listPageHelper(width int) string {
	return style.HelperView([]style.HelperContent{
		{Key: "<", Action: "menu"},
		{Key: "↑/↓", Action: "navigate"},
		{Key: "Enter", Action: "select"},
		{Key: "q", Action: "quit"},
		{Key: "?", Action: "toggle helper"},
	}, width, style.NormalView)
}
