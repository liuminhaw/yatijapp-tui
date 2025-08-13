package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type targetViewPage struct {
	cfg    config
	uuid   string
	target data.Target

	viewport viewport.Model

	width  int
	height int

	spinner      spinner.Model
	loading      bool
	confirmation *style.ConfirmCheck

	msg   string
	error error

	prev tea.Model // Previous model for navigation
}

func newTargetViewPage(
	cfg config,
	uuid string,
	termSize, vpSize style.ViewSize,
	prev tea.Model,
) targetViewPage {
	s := spinner.New()

	vp := viewport.New(vpSize.Width, vpSize.Height)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colors.BorderDimFg)

	return targetViewPage{
		cfg:      cfg,
		uuid:     uuid,
		viewport: vp,
		width:    termSize.Width,
		height:   termSize.Height,
		spinner:  s,
		loading:  true,
		prev:     prev,
	}
}

func (t targetViewPage) loadTarget(uuid string, msg string) tea.Cmd {
	serverURL := t.cfg.serverURL

	return func() tea.Msg {
		t.clearMsg()

		return loadTarget(serverURL, uuid, msg)
	}
}

func (t targetViewPage) deleteTarget(uuid, msg string) tea.Cmd {
	serverURL := t.cfg.serverURL

	return func() tea.Msg {
		t.clearMsg()
		if err := data.DeleteTarget(serverURL, uuid); err != nil {
			return apiResponseErrorMsg(err)
		}

		return targetDeletedMsg(msg)
	}
}

func (t *targetViewPage) setConfirmation() {
	t.confirmation = &style.ConfirmCheck{
		Prompt:  fmt.Sprintf("Proceed to delete target \"%s\"?", t.target.Title),
		Warning: "All activities and sessions under this target will be deleted as well.",
	}
}

func (t targetViewPage) Init() tea.Cmd {
	return tea.Batch(t.spinner.Tick, t.loadTarget(t.uuid, ""))
}

func (t targetViewPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
	case apiSuccessResponseMsg:
		t.loading = true
		return t, t.loadTarget(t.uuid, msg.msg)
	case tea.KeyMsg:
		if t.confirmation != nil {
			switch msg.String() {
			case "y":
				t.confirmation = nil
				return t, t.deleteTarget(t.target.UUID, "Target deleted successfully.")
			case "n":
				t.confirmation = nil
				return t, nil
			default:
				return t, nil
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return t, tea.Quit
		case "<":
			return t, switchToPreviousCmd(t.prev)
		case "e":
			return t, switchToTargetEditCmd(t.target)
		case "d":
			t.setConfirmation()
			return t, nil
		}
	case getTargetLoadedMsg:
		t.target = data.Target(msg.target)
		if err := t.renderViewport(); err != nil {
			// return t, func() tea.Msg { return getTargetLoadingErrorMsg(err) }
			return t, func() tea.Msg { return apiResponseErrorMsg(err) }
		}
		t.msg = msg.msg
		t.loading = false
	case targetDeletedMsg:
		return t, switchToPreviousCmd(t.prev)
	case apiResponseErrorMsg:
		t.error = msg
		t.loading = false
	case spinner.TickMsg:
		t.spinner, cmd = t.spinner.Update(msg)
		return t, cmd
	}

	t.viewport, cmd = t.viewport.Update(msg)
	return t, cmd
}

func (t targetViewPage) View() string {
	var title string
	if t.msg == "" {
		title = style.TitleBarView("Target Details", 80, false)
	} else {
		title = style.TitleBarView(t.msg, 80, true)
	}

	if t.error != nil {
		container := lipgloss.JoinVertical(
			lipgloss.Center,
			title,
			style.ErrorView(style.ViewSize{Width: 80, Height: 10}, t.error, true),
		)

		return style.ContainerStyle(t.width, container, 5).Render(container)
	}

	if t.loading {
		container := style.LoadingView(
			&t.spinner,
			"Target Details",
			style.ViewSize{Width: 80, Height: 10},
		)

		return style.ContainerStyle(t.width, container, 5).Render(container)
	}

	if t.confirmation != nil {
		container := t.confirmation.View(
			"Confirm Deletion",
			style.ViewSize{Width: 80, Height: 10},
		)
		return style.ContainerStyle(t.width, container, 5).Render(container)
	}

	helperView := style.HelperView([]style.HelperContent{
		{Key: "<", Action: "back"},
		{Key: "↑/↓", Action: "scroll"},
		{Key: "e", Action: "edit"},
		{Key: "d", Action: "delete"},
		{Key: "q", Action: "quit"},
	}, 80, style.NormalView)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		t.viewport.View(),
		helperView,
	)

	return style.ContainerStyle(t.width, container, 5).Render(container)
}

func (t targetViewPage) viewportContent() string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# %s\n", t.target.Title))
	content.WriteString(t.target.Description + "\n\n")

	content.WriteString("## Status\n")
	content.WriteString(fmt.Sprintf("%s\n\n", strings.ToUpper(t.target.Status)))

	content.WriteString("## Timestamp\n")
	if t.target.DueDate.Valid {
		content.WriteString(
			fmt.Sprintf("- **Due Date:** %s\n", t.target.DueDate.Time.Format("2006-01-02")),
		)
	} else {
		content.WriteString("- **Due Date:** --\n")
	}
	content.WriteString(
		fmt.Sprintf("- **Created At:** %s\n", t.target.CreatedAt.Format("2006-01-02 15:04:05")),
	)
	content.WriteString(
		fmt.Sprintf("- **Updated At:** %s\n", t.target.UpdatedAt.Format("2006-01-02 15:04:05")),
	)
	content.WriteString("\n")

	content.WriteString("## Notes\n---\n")
	if t.target.Notes == "" {
		content.WriteString("(Empty Note)")
	} else {
		content.WriteString(t.target.Notes)
	}

	return content.String()
}

func (t *targetViewPage) renderViewport() error {
	const glamourGutter = 2
	glamourRenderWidth := t.viewport.Width - glamourGutter - t.viewport.Style.GetHorizontalFrameSize()

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourRenderWidth),
		glamour.WithEmoji(),
	)
	if err != nil {
		return err
	}

	str, err := renderer.Render(t.viewportContent())
	if err != nil {
		return err
	}
	t.viewport.SetContent(str)

	return nil
}

func (t *targetViewPage) clearMsg() {
	t.msg = ""
}
