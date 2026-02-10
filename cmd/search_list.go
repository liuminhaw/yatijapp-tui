package main

import (
	"errors"
	"log/slog"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
	"github.com/liuminhaw/yatijapp-tui/pkg/strview"
)

type searchListPage struct {
	cfg    config
	helper bool

	selection recordsSelection
	hooks     listHooks

	width  int
	height int

	spinner spinner.Model
	loading bool

	prev tea.Model

	popupModels []tea.Model
	popup       string

	error error
}

func newSearchListPage(
	cfg config,
	q string,
	termSize style.ViewSize,
	prev tea.Model,
) searchListPage {
	s := searchListPage{
		cfg:       cfg,
		selection: newRecordsSelection(10),
		hooks: listHooks{
			loadAll: loadAllRecords,
			load:    loadRecord,
		},
		width:       termSize.Width,
		height:      termSize.Height,
		spinner:     spinner.New(spinner.WithSpinner(spinner.Line)),
		loading:     true,
		prev:        prev,
		popupModels: []tea.Model{},
	}
	s.selection.query["search"] = q

	return s
}

func (s searchListPage) Init() tea.Cmd {
	return tea.Batch(
		s.spinner.Tick,
		s.hooks.loadAll(
			data.ListRequestInfo{
				ServerURL:    s.cfg.apiEndpoint,
				SrcUUID:      "",
				QueryStrings: s.selection.query,
			},
			"", "searchList", s.cfg.authClient,
		),
	)
}

func (s searchListPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return s, tea.Quit
		case "up", "k":
			return s, s.selection.prev()
		case "down", "j":
			return s, s.selection.next()
		case "right", "l":
			return s, s.selection.nextPage()
		case "left", "h":
			return s, s.selection.prevPage()
		case "enter":
			if !s.selection.hasRecords() {
				break
			}
			s.cfg.logger.Info("Enter pressed in search list page")
			selected := s.selection.current()
			s.cfg.logger.Info("Actual type of selected record", slog.String("type", string(selected.GetActualType())))
			switch selected.GetActualType() {
			case data.RecordTypeTarget, data.RecordTypeAction:
				return s, switchToRecordsCmd(selected)
			}
		case "e":
			selected := s.selection.current()
			if selected == nil {
				break
			}

			s.loading = true
			return s, s.hooks.load(
				s.cfg.apiEndpoint, selected.GetUUID(), "", selected.GetActualType(), s.cfg.authClient,
			)
		case "<":
			if s.error != nil {
				s.error = nil
				return s, nil
			}
			return s, switchToPreviousCmd(s.prevPage())
		case "?":
			s.helper = !s.helper
			if s.helper {
				s.helperPopup(20)
			} else {
				s.popup = ""
			}
		}
	case allRecordsLoadedMsg:
		if msg.src != "searchList" {
			break
		}
		s.selection.setRecords(msg, s.cfg.logger)
		s.loading = false
	case loadMoreRecordsMsg:
		s.loading = true
		return s, s.hooks.loadAll(
			data.ListRequestInfo{
				ServerURL:    s.cfg.apiEndpoint,
				QueryStrings: s.selection.query,
				Events:       []string{msg.direction},
			},
			"", "searchList", s.cfg.authClient,
		)
	case getRecordLoadedMsg:
		s.loading = false
		return s, switchToEditCmd(s.selection.current().GetActualType(), msg.record)
	case data.UnauthorizedApiDataErr:
		s.cfg.logger.Error(
			msg.Error(),
			slog.Int("status", msg.Status),
			slog.String("occurence", "search list page"),
			slog.String("type", "record"),
		)
		s.loading = false
		return s, switchToMenuCmd
	case data.UnexpectedApiDataErr:
		s.cfg.logger.Error(
			msg.Error(),
			slog.String("occurence", "search list page"),
			slog.String("type", "record"),
		)
		s.error = errors.New(msg.Msg)
		s.loading = false
	case error:
		s.cfg.logger.Error(msg.Error(), slog.String("occurence", "search list page"))
		s.error = msg
		s.loading = false
	case spinner.TickMsg:
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd
	}

	return s, nil
}

func (s searchListPage) View() string {
	if s.loading {
		container := style.LoadingView(
			&s.spinner,
			"Loading list",
			style.ViewSize{Width: viewWidth, Height: 10},
		)

		return style.ContainerStyle(s.width, container, 5).Render(container)
	}

	title := style.TitleBarView([]string{"Search Results"}, viewWidth, false)

	if s.error != nil {
		return style.FullPageErrorView(
			title,
			s.width,
			style.ViewSize{Width: viewWidth, Height: 16},
			s.error,
			[]style.HelperContent{{Key: "<", Action: "back"}},
		)
	}

	helperView := s.listPageHelper(viewWidth)

	var content strings.Builder
	start, end := s.selection.p.GetSliceBounds(len(s.selection.records))
	for i, record := range s.selection.records[start:end] {
		content.WriteString(
			record.ListItemView(false, i+start == s.selection.selected, viewWidth),
		)
	}

	for i := len(s.selection.records[start:end]); i < s.selection.p.PerPage; i++ {
		content.WriteString("\n")
	}
	contentView := lipgloss.NewStyle().Render(content.String())

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		contentView,
		s.selection.view(),
		helperView,
	)

	if s.popup != "" {
		overlayX := lipgloss.Width(container)/2 - lipgloss.Width(s.popup)/2
		overlayY := lipgloss.Height(container)/2 - lipgloss.Height(s.popup)/2
		container = strview.PlaceOverlay(overlayX, overlayY, s.popup, container)
	}
	return style.ContainerStyle(s.width, container, 5).Render(container)
}

func (s *searchListPage) helperPopup(width int) {
	items := map[string]string{
		"<":       "Back",
		"↑/↓/←/→": "Navigate",
		"q":       "Quit",
		"v":       "View",
		"e":       "Edit",
		"m":       "Menu",
		"<C-/>":   "Search all",
		"?":       "Toggle helper",
	}
	order := []string{"<", "↑/↓/←/→", "v", "e", "m", "<C-/>", "q", "?"}

	if s.selection.records[s.selection.selected].GetActualType() != data.RecordTypeSession {
		items["Enter"] = "Select"
		order = slices.Insert(order, 2, "Enter")
	}

	s.popup = style.FullHelpView([]style.FullHelpContent{
		{
			Title:        "Key Maps",
			Items:        items,
			Order:        order,
			KeyHighlight: true,
		},
	}, width)
}

func (s searchListPage) listPageHelper(width int) string {
	content := []style.HelperContent{
		{Key: "<", Action: "back"},
		{Key: "↑/↓", Action: "navigate"},
		{Key: "?", Action: "toggle help"},
	}

	if len(s.selection.records) > 0 {
		if s.selection.records[s.selection.selected].GetActualType() != data.RecordTypeSession {
			content = slices.Insert(content, 2, style.HelperContent{Key: "Enter", Action: "select"})
		}
	}

	return style.HelperView(content, width)
}

func (s searchListPage) prevPage() tea.Model {
	if v, ok := s.prev.(listPage); ok {
		v.popupModels = []tea.Model{}
		v.popup = ""
		v.selectionSearchClear()

		return v
	} else if v, ok := s.prev.(menuPage); ok {
		v.popupModels = []tea.Model{}
		v.popup = ""

		return v
	}

	panic("previous page is not supported")
}
