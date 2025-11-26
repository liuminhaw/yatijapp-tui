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
	cfg    config
	helper bool
	uuid   string

	record     yatijappRecord
	hooks      viewHooks
	recordType data.RecordType

	viewport viewport.Model
	fullView bool

	width  int
	height int

	spinner spinner.Model
	loading bool

	msg   string
	error error
	prev  tea.Model // Previous model for navigation

	popupModels []tea.Model
	popup       string
}

func newViewPage(
	cfg config,
	uuid string,
	termSize, vpSize style.ViewSize,
	prev tea.Model,
) viewPage {
	vp := viewport.New(vpSize.Width, vpSize.Height)
	vp.Style = style.BorderStyle["focused"]

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

	var popupModel tea.Model
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "?":
			v.clearMsg()
			v.helper = !v.helper
			if v.helper {
				v.helperPopup(30)
			} else {
				v.popup = ""
			}
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		if err := v.viewportDisplay(); err != nil {
			return v, internalErrorCmd("failed to adjust view page size", err)
		}
	case tea.KeyMsg:
		if v.popup != "" {
			v.viewport.Style = style.BorderStyle["dimmed"]
			break
		} else {
			v.viewport.Style = style.BorderStyle["focused"]
		}
		switch msg.String() {
		case "ctrl+f":
			v.fullView = !v.fullView
			if err := v.viewportDisplay(); err != nil {
				return v, internalErrorCmd("failed to toggle full view mode", err)
			}
		case "ctrl+e":
			note, err := data.NewTempNote("view")
			if err != nil {
				return v, internalErrorCmd("error occurs when viewing note in editor", err)
			}
			if err := note.Write([]byte(v.viewportContent())); err != nil {
				return v, internalErrorCmd("error occurs when viewing note in editor", err)
			}
			if err := note.ReadOnly(); err != nil {
				return v, internalErrorCmd("error occurs when viewing note in editor", err)
			}
			return v, model.OpenEditor(note.Path())
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
			popupModel = model.NewAlert(
				"Confirm Deletion", "confirmation", prompts, warnings, 60,
				map[string]tea.Cmd{"confirm": deleteCmd, "cancel": cancelPopupCmd},
			)
			v.popupModels = append(v.popupModels, popupModel)
			v.popup = v.popupModels[len(v.popupModels)-1].View()
			return v, nil
		}
	case cancelPopupMsg:
		v.popupModels = v.popupModels[:len(v.popupModels)-1]
		if len(v.popupModels) == 0 {
			v.popup = ""
		} else {
			v.popup = v.popupModels[len(v.popupModels)-1].View()
		}
		v.clearMsg()
		return v, nil
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

	cmds = append(cmds, cmd)
	if len(v.popupModels) > 0 {
		lastIndex := len(v.popupModels) - 1
		v.popupModels[lastIndex], cmd = v.popupModels[lastIndex].Update(msg)
		v.popup = v.popupModels[lastIndex].View()
		cmds = append(cmds, cmd)
	}

	v.viewport, cmd = v.viewport.Update(msg)

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
		{Key: "?", Action: "modes"},
	}, viewWidth)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		v.viewport.View(),
		helperView,
	)

	if v.popup != "" {
		overlayX := lipgloss.Width(container)/2 - lipgloss.Width(v.popup)/2
		overlayY := lipgloss.Height(container)/2 - lipgloss.Height(v.popup)/2
		container = strview.PlaceOverlay(overlayX, overlayY, v.popup, container)
	}

	return style.ContainerStyle(v.width, container, 5).Render(container)
}

func (v viewPage) viewportContent() string {
	var content strings.Builder

	content.WriteString("# " + v.record.GetTitle() + "\n")
	description := v.record.GetDescription()
	if description != "" {
		content.WriteString(description + "\n\n")
	} else {
		content.WriteString("\n")
	}

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

func (v *viewPage) helperPopup(width int) {
	items := map[string]string{
		"<C-f>": "Toggle full screen",
		"<C-e>": "Open in editor",
		"?":     "Toggle mode helper",
	}
	order := []string{"<C-f>", "<C-e>", "?"}

	v.popup = style.FullHelpView([]style.FullHelpContent{
		{
			Title:        "View Modes",
			Items:        items,
			Order:        order,
			KeyHighlight: true,
		},
	}, width)
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

func (v *viewPage) viewportDisplay() error {
	if v.width <= viewWidth || v.height <= 28 || !v.fullView {
		v.viewport.Width = viewWidth
		v.viewport.Height = 20
		v.fullView = false
	} else {
		v.viewport.Width = v.width
		v.viewport.Height = v.height - 8
	}

	return v.renderViewport()
}
