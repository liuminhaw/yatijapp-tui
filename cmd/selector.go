package main

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type selectorHooks struct {
	loadAll func(serverURL, srcUUID, msg string, client *authclient.AuthClient) tea.Cmd
}

type selectorPage struct {
	cfg config

	selection recordsSelection
	// selection model.SelectorModel
	// text      *components.TextComponent

	hooks      selectorHooks
	recordType data.RecordType

	width  int
	height int

	spinner spinner.Model
	loading bool

	error error

	prev tea.Model
}

func newSelectorPage(cfg config, termSize style.ViewSize, prev tea.Model) selectorPage {
	return selectorPage{
		cfg: cfg,
		// paginator: style.NewPaginator(10),
		selection: recordsSelection{
			p: style.NewPaginator(10),
		},
		// selection: model.NewSelectorModel([]data.YatijappRecord{}, 10, 70),
		// text:      components.NewText(""),
		width:   termSize.Width,
		height:  termSize.Height,
		spinner: spinner.New(spinner.WithSpinner(spinner.Line)),
		loading: true,
		prev:    prev,
	}
}

func newTargetSelectorPage(cfg config, termSize style.ViewSize, prev tea.Model) selectorPage {
	page := newSelectorPage(cfg, termSize, prev)
	page.recordType = data.RecordTypeTarget
	page.hooks = selectorHooks{
		loadAll: loadAllTargets,
	}
	return page
}

func (p selectorPage) Init() tea.Cmd {
	return tea.Batch(p.spinner.Tick, p.hooks.loadAll(p.cfg.serverURL, "", "", p.cfg.authClient))
}

func (p selectorPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return p, tea.Quit
		case "up", "k":
			p.selection.prev()
		case "down", "j":
			p.selection.next()
		case "right", "l":
			p.selection.nextPage()
		case "left", "h":
			p.selection.prevPage()
		case "<":
			return p, switchToPreviousCmd(p.prev)
		case "enter":
			if len(p.selection.records) == 0 {
				return p, nil
			}
			selected := p.selection.records[p.selection.selected]
			return p, selectorSelectedCmd(p.prev, selected.GetTitle(), selected.GetUUID(), p.recordType)
		}
	case allRecordsLoadedMsg:
		p.selection.records = msg.records
		p.selection.p.SetTotalPages(len(p.selection.records))
		p.loading = false
	case data.UnauthorizedApiDataErr:
		p.cfg.logger.Error(
			msg.Error(),
			slog.Int("status", msg.Status),
			slog.String("action", "load records"),
			slog.String("type", string(p.recordType)),
		)
		p.loading = false
		return p, switchToMenuCmd
	case data.UnexpectedApiDataErr:
		p.cfg.logger.Error(
			msg.Error(),
			slog.String("action", "load records"),
			slog.String("type", string(p.recordType)),
		)
		p.error = errors.New(msg.Msg)
		p.loading = false
	case error:
		p.cfg.logger.Error(msg.Error(), slog.String("occurence", "selector page"))
		p.error = msg
		p.loading = false
	case spinner.TickMsg:
		p.spinner, cmd = p.spinner.Update(msg)
		return p, cmd
	}

	p.selection.p, cmd = p.selection.p.Update(msg)
	return p, cmd
}

func (p selectorPage) View() string {
	if p.loading {
		container := style.LoadingView(
			&p.spinner,
			"Loading "+strings.ToLower(string(p.recordType))+"s",
			style.ViewSize{Width: viewWidth, Height: 10},
		)

		return style.ContainerStyle(p.width, container, 5).Render(container)
	}

	title := style.TitleBarView([]string{"Select " + string(p.recordType)}, viewWidth, false)

	if p.error != nil {
		return style.FullPageErrorView(
			title,
			p.width,
			style.ViewSize{Width: viewWidth, Height: 10},
			p.error,
			[]style.HelperContent{{Key: "<", Action: "back"}},
		)
	}

	var content strings.Builder
	start, end := p.selection.p.GetSliceBounds(len(p.selection.records))
	for i, record := range p.selection.records[start:end] {
		content.WriteString(selectionView(record, i+start == p.selection.selected, 70))
	}

	for i := len(p.selection.records[start:end]); i < p.selection.p.PerPage; i++ {
		content.WriteString("\n")
	}
	content.WriteString(
		lipgloss.NewStyle().Width(70).AlignHorizontal(lipgloss.Center).Render(p.selection.p.View()),
	)

	helperView := style.HelperView([]style.HelperContent{
		{Key: "<", Action: "back"},
		// {Key: "↑/↓/←/→", Action: "navigate"},
		{Key: "arrow keys", Action: "navigate"},
		{Key: "enter", Action: "select"},
		{Key: "q", Action: "quit"},
	}, viewWidth, style.NormalView)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		lipgloss.NewStyle().
			Height(15).
			Margin(1).
			Padding(1, 2, 0).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.BorderDimFg).
			Render(content.String()),
		helperView,
	)

	return style.ContainerStyle(p.width, container, 5).Render(container)
}

func selectionView(record yatijappRecord, selected bool, width int) string {
	var stringBuilder strings.Builder

	if selected {
		stringBuilder.WriteString(
			lipgloss.NewStyle().Bold(true).Foreground(colors.DocumentText).Render("▸ ") +
				style.ChoicesStyle["list"].Choice.
					Width(width-3).
					Render(" "+record.GetTitle()) + "\n",
		)

		hr := lipgloss.NewStyle().Foreground(colors.HelperText).
			Render("  " + strings.Repeat("─", width-3))
		stringBuilder.WriteString(hr + "\n")

		var description string
		if record.GetDescription() == "" {
			description = "---"
		} else {
			description = strings.TrimSpace(record.GetDescription())
		}
		stringBuilder.WriteString(
			style.ChoicesStyle["list"].ChoiceContent.
				Width(width).
				Render("  "+description) + "\n",
		)
		stringBuilder.WriteString(hr + "\n")
	} else {
		stringBuilder.WriteString(
			style.ChoicesStyle["default"].Choices.Width(width).Render("▫ "+record.GetTitle()) + "\n",
		)
	}

	return stringBuilder.String()
}
