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

	parentUUID string
	selection  recordsSelection
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

func newSelectorPage(
	cfg config,
	termSize style.ViewSize,
	parentUUID string,
	prev tea.Model,
) selectorPage {
	return selectorPage{
		cfg: cfg,
		selection: recordsSelection{
			p: style.NewPaginator(6),
		},
		parentUUID: parentUUID,
		width:      termSize.Width,
		height:     termSize.Height,
		spinner:    spinner.New(spinner.WithSpinner(spinner.Line)),
		loading:    true,
		prev:       prev,
	}
}

func newTargetSelectorPage(cfg config, termSize style.ViewSize, prev tea.Model) selectorPage {
	page := newSelectorPage(cfg, termSize, "", prev)
	page.recordType = data.RecordTypeTarget
	page.hooks = selectorHooks{loadAll: loadAllTargets}
	return page
}

func newActionSelectorPage(
	cfg config,
	termSize style.ViewSize,
	targetUUID string,
	prev tea.Model,
) selectorPage {
	page := newSelectorPage(cfg, termSize, targetUUID, prev)
	page.recordType = data.RecordTypeAction
	page.hooks = selectorHooks{loadAll: loadAllActions}

	return page
}

func (p selectorPage) Init() tea.Cmd {
	var cmd tea.Cmd
	if p.parentUUID == "" && p.recordType != data.RecordTypeTarget {
		cmd = func() tea.Msg {
			return allRecordsLoadedMsg{}
		}
	} else {
		cmd = p.hooks.loadAll(p.cfg.serverURL, p.parentUUID, "", p.cfg.authClient)
	}
	return tea.Batch(p.spinner.Tick, cmd)
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
	var content strings.Builder

	if p.loading {
		msg := style.Document.NormalDim.Bold(true).Render("loading...")
		p.spinner.Style = style.Document.Highlight
		content.WriteString(p.spinner.View() + " " + msg)

		container := lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.NewStyle().
				Width(78).
				Height(10).
				Align(lipgloss.Center, lipgloss.Center).
				Render(content.String()),
			style.Document.Primary.Render("<")+style.Document.Normal.Render(" back\n"),
		)
		return lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.BorderDimFg).
			Render(container)
	}

	if p.error != nil {
		container := lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.NewStyle().
				Render(
					style.ErrorView(style.ViewSize{Width: 72, Height: 10}, p.error, nil),
				),
			style.Document.Primary.Render("<")+style.Document.Normal.Render(" back\n"),
		)
		return lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.BorderDimFg).
			Render(container)
	}

	title := style.Document.Secondary.Bold(true).
		Render("Select " + strings.ToLower(string(p.recordType)))

	start, end := p.selection.p.GetSliceBounds(len(p.selection.records))
	for i, record := range p.selection.records[start:end] {
		content.WriteString(selectionView(record, i+start == p.selection.selected, 72))
	}

	for i := len(p.selection.records[start:end]); i < p.selection.p.PerPage; i++ {
		content.WriteString("\n")
	}

	helper := lipgloss.StyleRanges(
		"< back    arrows navigate    enter select\n",
		lipgloss.Range{Start: 0, End: 1, Style: style.Document.Primary},
		lipgloss.Range{Start: 2, End: 6, Style: style.Document.Normal},
		lipgloss.Range{Start: 10, End: 16, Style: style.Document.Primary},
		lipgloss.Range{Start: 16, End: 25, Style: style.Document.Normal},
		lipgloss.Range{Start: 28, End: 34, Style: style.Document.Primary},
		lipgloss.Range{Start: 35, End: 41, Style: style.Document.Normal},
	)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		lipgloss.NewStyle().
			Height(7).
			Margin(1, 0, 0).
			Padding(0, 2, 0).
			Render(content.String()),
		lipgloss.NewStyle().
			Width(70).
			AlignHorizontal(lipgloss.Center).
			Render(p.selection.p.View()),
		helper,
	)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colors.TextMuted).
		Render(container)
}

func selectionView(record yatijappRecord, selected bool, width int) string {
	var builder strings.Builder

	if selected {
		builder.WriteString(
			style.StatusTextStyle(record.GetStatus()).MarginLeft(1).Render("∎") +
				style.ChoicesStyle["list"].Choice.Width(width-2).
					Margin(0, 1, 0, 0).
					Padding(0, 1, 0, 1).
					Render(record.GetTitle()) + "\n",
		)
	} else {
		builder.WriteString(
			style.StatusTextStyle(record.GetStatus()).MarginLeft(1).Render("∎") +
				lipgloss.NewStyle().Width(width).Padding(0, 1).Render(
					style.ChoicesStyle["list"].Choices.Render(record.GetTitle()),
				) + "\n",
		)
	}

	return builder.String()
}
