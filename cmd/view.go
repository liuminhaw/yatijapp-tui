package main

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/model"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
	"github.com/liuminhaw/yatijapp-tui/pkg/strview"
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

	spinner spinner.Model
	loading bool

	msg   string
	error error
	prev  tea.Model // Previous model for navigation

	popupModel tea.Model
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
		BorderForeground(colors.TextMuted)

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

func newActionViewPage(
	cfg config,
	uuid string,
	termSize, vpSize style.ViewSize,
	prev tea.Model,
) viewPage {
	page := newViewPage(cfg, uuid, termSize, vpSize, prev)
	page.recordType = data.RecordTypeAction
	page.hooks.load = loadAction
	page.hooks.delete = deleteAction
	return page
}

func newSessionViewPage(
	cfg config,
	uuid string,
	termSize, vpSize style.ViewSize,
	prev tea.Model,
) viewPage {
	page := newViewPage(cfg, uuid, termSize, vpSize, prev)
	page.recordType = data.RecordTypeSession
	page.hooks.load = loadSession
	page.hooks.delete = deleteSession
	return page
}

func (v viewPage) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.hooks.load(v.cfg.apiEndpoint, v.uuid, "", v.cfg.authClient))
}

func (v viewPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return v, tea.Quit
		case "<":
			return v, switchToPreviousCmd(v.prevPage())
		case "e":
			return v, switchToEditCmd(v.recordType, v.record)
		case "d":
			if v.record == nil {
				panic("view page item is nil in delete")
			}
			var prompts, warnings []string
			switch v.record.GetActualType() {
			case data.RecordTypeTarget:
				prompts = []string{"Proceed to delete target \"" + v.record.GetTitle() + "\"?"}
				warnings = []string{"All actions and sessions under this target will be deleted as well."}
			case data.RecordTypeAction:
				prompts = []string{"Proceed to delete action \"" + v.record.GetTitle() + "\"?"}
				warnings = []string{"All sessions under this action will be deleted as well."}
			case data.RecordTypeSession:
				if v.record.GetStatus() == "completed" {
					prompts = []string{
						"Proceed to delete session",
						"\"" + v.record.GetTitle() + "\"?",
					}
				} else {
					prompts = []string{"Proceed to delete session \"" + v.record.GetTitle() + "\"?"}
				}
			}
			deleteCmd := v.hooks.delete(v.cfg.apiEndpoint, v.uuid, v.cfg.authClient)
			v.popupModel = model.NewAlert(
				"Confirm Deletion",
				"confirmation",
				prompts,
				warnings,
				60,
				map[string]tea.Cmd{"confirm": deleteCmd, "cancel": cancelPopupCmd},
			)
			return v, nil
		}
	case cancelPopupMsg:
		v.popupModel = nil
		v.clearMsg()
	case apiSuccessResponseMsg:
		v.loading = false
		return v, v.hooks.load(v.cfg.apiEndpoint, v.uuid, msg.msg, v.cfg.authClient)
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
	cmds = append(cmds, cmd)
	if v.popupModel != nil {
		v.popupModel, cmd = v.popupModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
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

	helperView := style.HelperView([]style.HelperContent{
		{Key: "<", Action: "back"},
		{Key: "↑/↓", Action: "scroll"},
		{Key: "e", Action: "edit"},
		{Key: "d", Action: "delete"},
		{Key: "q", Action: "quit"},
	}, viewWidth)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		v.viewport.View(),
		helperView,
	)

	if v.popupModel != nil {
		overlayX := lipgloss.Width(container)/2 - lipgloss.Width(v.popupModel.View())/2
		overlayY := lipgloss.Height(container)/2 - lipgloss.Height(v.popupModel.View())/2
		container = strview.PlaceOverlay(overlayX, overlayY, v.popupModel.View(), container)
	}

	return style.ContainerStyle(v.width, container, 5).Render(container)
}

func (v viewPage) viewportContent() string {
	var content strings.Builder

	content.WriteString("# " + v.record.GetTitle() + "\n")
	content.WriteString(v.record.GetDescription() + "\n\n")

	switch v.recordType {
	case data.RecordTypeAction:
		content.WriteString("## Upstream\n")
		content.WriteString(
			"- **Target:** " + v.record.GetParentsTitle()[data.RecordTypeTarget] + "\n\n",
		)
	case data.RecordTypeSession:
		content.WriteString("## Upstream\n")
		content.WriteString(
			"- **Target:** " + v.record.GetParentsTitle()[data.RecordTypeTarget] + "\n" +
				"- **Action:** " + v.record.GetParentsTitle()[data.RecordTypeAction] + "\n\n",
		)
	}

	content.WriteString("## Status\n")
	content.WriteString(strings.ToUpper(v.record.GetStatus()) + "\n\n")

	due, valid := v.record.GetDueDate()
	content.WriteString("## Timestamp\n")
	if v.recordType != data.RecordTypeSession {
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
	}

	if v.recordType == data.RecordTypeSession {
		session := v.record.(data.Session)
		content.WriteString(
			"- **Starts At:**  " + session.StartsAt.Format("2006-01-02 15:04:05") + "\n",
		)
		if session.EndsAt.Valid {
			content.WriteString(
				"- **Ends At:**    " + session.EndsAt.Time.Format("2006-01-02 15:04:05") + "\n",
			)
			content.WriteString(
				"- **Duration:**   " + session.EndsAt.Time.Sub(session.StartsAt).
					Truncate(time.Second).
					String() +
					"\n\n",
			)
		} else {
			content.WriteString("- **Ends At:**    --\n\n")
		}
	}

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
		parentType := p.recordType.GetParentType()
		if parentType == "" {
			return p.prev
		}

		parentUUID := v.src.UUID(parentType)
		if parentUUID != "" && parentUUID != p.record.GetParentsUUID()[parentType] {
			p.cfg.logger.Info(
				fmt.Sprintf(
					"before parent uuid: %s, parent title: %s",
					parentUUID,
					v.src.Title(parentType),
				),
			)
			v.src[parentType] = data.RecordParent{
				UUID:  p.record.GetParentsUUID()[parentType],
				Title: p.record.GetParentsTitle()[parentType],
			}
			p.cfg.logger.Info(
				fmt.Sprintf(
					"after parent uuid: %s, parent title: %s",
					v.src.UUID(parentType),
					v.src.Title(parentType),
				),
			)
			return v
		}
	}

	return p.prev
}
