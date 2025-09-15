package main

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type viewHooks struct {
	load   func(serverURL, uuid, msg string, client *authclient.AuthClient) tea.Cmd
	delete func(serverURL, uuid string, client *authclient.AuthClient) tea.Cmd
}

type viewPage struct {
	cfg  config
	uuid string

	record     yatijappRecord
	hooks      viewHooks
	recordType data.RecordType

	viewport viewport.Model

	width  int
	height int

	spinner      spinner.Model
	loading      bool
	confirmation *style.ConfirmCheck

	msg   string
	error error
	prev  tea.Model // Previous model for navigation
}

func newViewPage(
	cfg config,
	uuid string,
	termSize, vpSize style.ViewSize,
	prev tea.Model,
) viewPage {
	vp := viewport.New(vpSize.Width, vpSize.Height)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colors.BorderDimFg)

	return viewPage{
		cfg:      cfg,
		uuid:     uuid,
		viewport: vp,
		width:    termSize.Width,
		height:   termSize.Height,
		spinner:  spinner.New(spinner.WithSpinner(spinner.Line)),
		loading:  true,
		prev:     prev,
	}
}

func newTargetViewPage2(
	cfg config,
	uuid string,
	termSize, vpSize style.ViewSize,
	prev tea.Model,
) viewPage {
	page := newViewPage(cfg, uuid, termSize, vpSize, prev)
	page.recordType = data.RecordTypeTarget
	page.hooks.load = loadTarget
	page.hooks.delete = deleteTarget
	return page
}

func newActivityViewPage(
	cfg config,
	uuid string,
	termSize, vpSize style.ViewSize,
	prev tea.Model,
) viewPage {
	page := newViewPage(cfg, uuid, termSize, vpSize, prev)
	page.recordType = data.RecordTypeActivity
	page.hooks.load = loadActivity
	page.hooks.delete = deleteActivity
	return page
}

func (v *viewPage) setConfirmation(prompt, warning string) {
	v.confirmation = &style.ConfirmCheck{Prompt: prompt, Warning: warning}
}

func (v viewPage) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.hooks.load(v.cfg.serverURL, v.uuid, "", v.cfg.authClient))
}

func (v viewPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
	case tea.KeyMsg:
		if v.confirmation != nil {
			switch msg.String() {
			case "y":
				v.confirmation = nil
				v.clearMsg()
				return v, v.hooks.delete(v.cfg.serverURL, v.uuid, v.cfg.authClient)
			case "n":
				v.confirmation = nil
				return v, nil
			default:
				return v, nil
			}
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return v, tea.Quit
		case "<":
			// return v, switchToPreviousCmd(v.prev)
			return v, switchToPreviousCmd(v.prevPage())
		case "e":
			return v, switchToEditCmd(v.recordType, v.record)
		case "d":
			if v.record == nil {
				panic("view page item is nil in delete")
			}
			var prompt, warning string
			switch v.record.GetActualType() {
			case data.RecordTypeTarget:
				prompt = "Proceed to delete target \"" + v.record.GetTitle() + "\"?"
				warning = "All activities and sessions under this target will be deleted as well."
			case data.RecordTypeActivity:
				prompt = "Proceed to delete activity \"" + v.record.GetTitle() + "\"?"
				warning = "All sessions under this activity will be deleted as well."
			}
			v.setConfirmation(prompt, warning)
			return v, nil
		}
	case apiSuccessResponseMsg:
		v.loading = false
		return v, v.hooks.load(v.cfg.serverURL, v.uuid, msg.msg, v.cfg.authClient)
	case getRecordLoadedMsg:
		v.record = msg.record
		if err := v.renderViewport(); err != nil {
			return v, internalErrorCmd("failed to render view page", err)
		}
		v.msg = msg.msg
		v.loading = false
	case recordDeletedMsg:
		return v, switchToPreviousCmd(v.prev)
	case internalErrorMsg:
		v.error = errors.New(msg.msg)
		v.loading = false
	case data.UnauthorizedApiDataErr:
		v.cfg.logger.Error(
			msg.Error(),
			slog.Int("status", msg.Status),
			slog.String("action", "load record"),
			slog.String("type", string(v.recordType)),
		)
		v.loading = false
		return v, switchToMenuCmd
	case data.UnexpectedApiDataErr:
		v.cfg.logger.Error(
			msg.Error(),
			slog.String("action", "load record"),
			slog.String("type", string(v.recordType)),
		)
		v.error = errors.New(msg.Msg)
		v.loading = false
	case error:
		v.cfg.logger.Error(
			msg.Error(),
			slog.String("occurence", "view page"),
		)
		v.error = msg
		v.loading = false
	case spinner.TickMsg:
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v viewPage) View() string {
	if v.loading {
		container := style.LoadingView(
			&v.spinner,
			"Loading Details",
			style.ViewSize{Width: viewWidth, Height: 10},
		)

		return style.ContainerStyle(v.width, container, 5).Render(container)
	}

	var title string
	if v.msg == "" {
		title = style.TitleBarView([]string{string(v.recordType) + " Details"}, viewWidth, false)
	} else {
		title = style.TitleBarView([]string{v.msg}, viewWidth, true)
	}

	if v.error != nil {
		return style.FullPageErrorView(
			title,
			v.width,
			style.ViewSize{Width: viewWidth, Height: 10},
			v.error,
			[]style.HelperContent{{Key: "<", Action: "back"}},
		)
	}

	if v.confirmation != nil {
		container := v.confirmation.View(
			"Confirm Deletion",
			style.ViewSize{Width: viewWidth, Height: 10},
		)
		return style.ContainerStyle(v.width, container, 5).Render(container)
	}

	helperView := style.HelperView([]style.HelperContent{
		{Key: "<", Action: "back"},
		{Key: "↑/↓", Action: "scroll"},
		{Key: "e", Action: "edit"},
		{Key: "d", Action: "delete"},
		{Key: "q", Action: "quit"},
	}, viewWidth, style.NormalView)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		v.viewport.View(),
		helperView,
	)

	return style.ContainerStyle(v.width, container, 5).Render(container)
}

func (v viewPage) viewportContent() string {
	var content strings.Builder

	content.WriteString("# " + v.record.GetTitle() + "\n")
	content.WriteString(v.record.GetDescription() + "\n\n")

	if v.recordType == data.RecordTypeActivity {
		content.WriteString("## Upstream\n")
		content.WriteString(
			"- **Target:** " + v.record.GetParentsTitle()[data.RecordTypeTarget] + "\n\n",
		)
	}

	content.WriteString("## Status\n")
	content.WriteString(strings.ToUpper(v.record.GetStatus()) + "\n\n")

	due, valid := v.record.GetDueDate()
	content.WriteString("## Timestamp\n")
	if valid {
		content.WriteString("- **Due Date:**: " + due.Format("2006-01-02") + "\n")
	} else {
		content.WriteString("- **Due Date:** --\n")
	}
	content.WriteString(
		"- **Created At:** " + v.record.GetCreatedAt().Format("2006-01-02 15:04:05") + "\n",
	)
	content.WriteString(
		"- **Updated At:** " + v.record.GetUpdatedAt().Format("2006-01-02 15:04:05") + "\n\n",
	)
	content.WriteString(
		"- **Last Active:** " + v.record.GetUpdatedAt().Format("2006-01-02 15:04:05") + "\n\n",
	)

	content.WriteString("## Notes\n---\n")
	note := v.record.GetNote()
	if note == "" {
		content.WriteString("(Empty Note)")
	} else {
		content.WriteString(note)
	}

	return content.String()
}

func (v *viewPage) renderViewport() error {
	const glamourGutter = 2
	glamourRenderWidth := v.viewport.Width - glamourGutter - v.viewport.Style.GetHorizontalFrameSize()

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourRenderWidth),
		glamour.WithEmoji(),
	)
	if err != nil {
		return err
	}

	str, err := renderer.Render(v.viewportContent())
	if err != nil {
		return err
	}
	v.viewport.SetContent(str)

	return nil
}

func (v *viewPage) clearMsg() {
	v.msg = ""
}

func (p viewPage) prevPage() tea.Model {
	if v, ok := p.prev.(listPage); ok {
		parentType, exists := p.recordType.GetParentType()
		if !exists {
			return p.prev
		}
		if v.src.uuid != "" && v.src.uuid != p.record.GetParentsUUID()[parentType] {
			p.cfg.logger.Info(
				fmt.Sprintf("before parent uuid: %s, parent title: %s", v.src.uuid, v.src.title),
			)
			v.src.uuid = p.record.GetParentsUUID()[parentType]
			v.src.title = p.record.GetParentsTitle()[parentType]
			p.cfg.logger.Info(
				fmt.Sprintf("after parent uuid: %s, parent title: %s", v.src.uuid, v.src.title),
			)
			return v
		}
	}

	return p.prev
}
