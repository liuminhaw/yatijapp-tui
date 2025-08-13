package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type targetListPage struct {
	cfg    config
	helper bool

	// Target list
	targets   []data.Target
	paginator paginator.Model
	selected  int

	width  int
	height int

	spinner      spinner.Model
	loading      bool
	confirmation *style.ConfirmCheck

	msg   string
	error error
}

func newTargetListPage(cfg config, width, height int) targetListPage {
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 10
	p.ActiveDot = lipgloss.NewStyle().Foreground(colors.DocumentText).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(colors.HelperTextDim).Render("•")

	s := spinner.New()

	return targetListPage{
		cfg:       cfg,
		paginator: p,
		width:     width,
		height:    height,
		spinner:   s,
		loading:   true,
	}
}

func (t targetListPage) loadTargetListPage(msg string) tea.Cmd {
	return func() tea.Msg {
		t.clearMsg()

		targets, err := data.ListTargets(t.cfg.serverURL)
		if err != nil {
			return apiResponseErrorMsg(err)
		}

		return targetListLoadedMsg{
			targets: targets,
			msg:     msg,
		}
	}
}

func (t targetListPage) loadTarget(uuid string, msg string) tea.Cmd {
	serverURL := t.cfg.serverURL

	return func() tea.Msg {
		t.clearMsg()

		return loadTarget(serverURL, uuid, msg)
	}
}

func (t targetListPage) deleteTarget(uuid, msg string) tea.Cmd {
	serverURL := t.cfg.serverURL

	return func() tea.Msg {
		t.clearMsg()
		err := data.DeleteTarget(serverURL, uuid)
		if err != nil {
			return apiResponseErrorMsg(err)
		}

		return targetDeletedMsg(msg)
	}
}

func (t *targetListPage) setConfirmation() {
	t.confirmation = &style.ConfirmCheck{
		Prompt:  fmt.Sprintf("Proceed to delete target \"%s\"?", t.targets[t.selected].Title),
		Warning: "All activities and sessions under this target will be deleted as well.",
	}
}

func (t targetListPage) Init() tea.Cmd {
	return tea.Batch(t.spinner.Tick, t.loadTargetListPage(""))
}

func (t targetListPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
	case switchToPreviousMsg:
		t.loading = true
		return t, t.loadTargetListPage("")
	case apiSuccessResponseMsg:
		t.loading = true
		return t, t.loadTargetListPage(msg.msg)
	case tea.KeyMsg:
		if t.confirmation != nil {
			switch msg.String() {
			case "y":
				t.confirmation = nil	
				return t, t.deleteTarget(t.targets[t.selected].UUID, "Target deleted successfully")
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
		case "up", "k":
			t.clearMsg()
			if t.selected > 0 {
				t.selected--
				start, _ := t.paginator.GetSliceBounds(len(t.targets))
				if t.selected < start {
					t.paginator.PrevPage()
				}
			}
		case "down", "j":
			t.clearMsg()
			if t.selected < len(t.targets)-1 {
				t.selected++
				_, end := t.paginator.GetSliceBounds(len(t.targets))
				if t.selected >= end {
					t.paginator.NextPage()
				}
			}
		case "right", "l":
			t.clearMsg()
			if t.paginator.OnLastPage() {
				break
			}
			t.paginator.NextPage()
			t.selected, _ = t.paginator.GetSliceBounds(len(t.targets))
		case "left", "h":
			t.clearMsg()
			if t.paginator.OnFirstPage() {
				break
			}
			t.paginator.PrevPage()
			t.selected, _ = t.paginator.GetSliceBounds(len(t.targets))
		case "<":
			return t, switchToMenuCmd
		case "?":
			t.clearMsg()
			t.helper = !t.helper
			// return t, switchToHelperListCmd
		case "v":
			return t, switchToTargetViewCmd(t.targets[t.selected].UUID)
		case "e":
			t.loading = true
			return t, t.loadTarget(t.targets[t.selected].UUID, "")
		case "d":
			t.setConfirmation()
			return t, confirmationCmd
		case "n":
			return t, switchToTargetCreateCmd
		case "ctrl+r":
			return t, switchToTargetsCmd
		}
	case targetListLoadedMsg:
		// t.targets = []data.Target(msg)
		t.msg = msg.msg
		t.targets = msg.targets
		t.paginator.SetTotalPages(len(t.targets))
		// status.promptStyle().Render(status.prompt),
		// status.contentStyle().Render(status.content),
		t.loading = false
	case getTargetLoadedMsg:
		t.loading = false
		return t, switchToTargetEditCmd(msg.target)
	case targetDeletedMsg:
		return t, t.loadTargetListPage(string(msg))
	case confirmationMsg:
		t.loading = false
	case apiResponseErrorMsg:
		t.error = msg
		t.loading = false
	case spinner.TickMsg:
		t.spinner, cmd = t.spinner.Update(msg)
		return t, cmd
	}

	t.paginator, cmd = t.paginator.Update(msg)
	return t, cmd
}

func (t targetListPage) View() string {
	var title string
	if t.msg == "" {
		title = style.TitleBarView("Targets", 80, false)
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
			"Targets",
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

	var sidebar string
	if !t.helper {
		sidebar = style.SideBarView([]style.SideBarContent{
			{
				Title: "Filters",
				Items: map[string]string{},
				Order: []string{"Status", "Search"},
			},
		}, 20)
	} else {
		sidebar = style.SideBarView([]style.SideBarContent{
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
		}, 20)
	}

	helperView := style.HelperView([]style.HelperContent{
		{Key: "<", Action: "menu"},
		{Key: "↑/↓", Action: "navigate"},
		{Key: "Enter", Action: "select"},
		{Key: "q", Action: "quit"},
		{Key: "?", Action: "toggle helper"},
	}, 80, style.NormalView)

	var content strings.Builder
	start, end := t.paginator.GetSliceBounds(len(t.targets))
	for i, target := range t.targets[start:end] {
		content.WriteString(target.ListItemView(i+start == t.selected, 57))
	}

	for i := len(t.targets[start:end]); i < t.paginator.PerPage; i++ {
		content.WriteString("\n")
	}
	content.WriteString(t.paginator.View())

	contentView := lipgloss.NewStyle().Height(lipgloss.Height(sidebar)).Render(content.String())

	container := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Margin(1, 0, 0).Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				sidebar,
				lipgloss.NewStyle().
					Margin(0, 1).
					Padding(0, 0, 0, 2).
					BorderStyle(lipgloss.NormalBorder()).
					BorderForeground(colors.DocumentTextDim).
					BorderLeft(true).
					Render(contentView),
			),
		),
		// msgView,
		helperView,
	)

	return style.ContainerStyle(t.width, container, 5).Render(container)
}

func (t *targetListPage) clearMsg() {
	t.msg = ""
}
