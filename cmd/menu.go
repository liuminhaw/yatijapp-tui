package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/components"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type menuPage struct {
	cfg config

	title string
	view  components.Radio

	authView   components.Radio
	unauthView components.Radio

	width  int
	height int

	spinner spinner.Model
	loading bool

	// msg string
	error error
}

func newMenuPage(cfg config, width, height int) menuPage {
	return menuPage{
		cfg:        cfg,
		title:      menuTitle(),
		authView:   menuView([]string{"Targets", "Activities", "Sessions"}),
		unauthView: menuView([]string{"Sign in", "Sign up"}),
		width:      width,
		height:     height,
		spinner:    spinner.New(spinner.WithSpinner(spinner.Meter)),
		loading:    true,
	}
}

func (m menuPage) loadLoginUser() tea.Cmd {
	return func() tea.Msg {
		user, err := data.GetCurrentUser(m.cfg.serverURL, m.cfg.authClient)
		if err != nil {
			return apiErrorResponseCmd(err)
		}

		return apiSuccessResponseMsg{
			msg:      user.Name,
			redirect: m,
		}
	}
}

func (m menuPage) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadLoginUser())
}

func (m menuPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			switch m.view.Selected() {
			case "Targets":
				return m, switchToTargetsCmd
			case "Activities":
				return m, switchToActivitiesCmd
			case "Sessions":
				return m, switchToSessionsCmd
			}
		}
	case apiSuccessResponseMsg:
		m.view = m.authView
		m.loading = false
	case data.LoadApiDataErr:
		switch msg.Status {
		case http.StatusUnauthorized, http.StatusForbidden:
			m.view = m.unauthView
		default:
			m.cfg.logger.Error(msg.Error(), slog.Int("status", msg.Status), slog.String("action", "load login user"))
			m.error = errors.New(msg.Msg)
		}
		m.loading = false
	case unexpectedApiResponseMsg:
		m.error = msg
		m.loading = false
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	m.view, cmd = m.view.Update(msg)

	return m, cmd
}

func (m menuPage) View() string {
	if m.error != nil {
		container := lipgloss.JoinVertical(
			lipgloss.Center,
			style.TitleBarView("Menu", viewWidth, false),
			style.ErrorView(
				style.ViewSize{Width: 80, Height: 10},
				m.error,
				[]style.HelperContent{{Key: "q", Action: "quit"}},
			),
		)

		return style.ContainerStyle(m.width, container, 5).Render(container)
	}

	if m.loading {
		msg := style.DocumentStyle.Normal.Bold(true).Render("Yatijapp")
		m.spinner.Style = style.DocumentStyle.Highlight
		container := lipgloss.NewStyle().
			Width(50).
			Height(10).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("%s %s", m.spinner.View(), msg))

		return style.ContainerStyle(m.width, container, 5).Render(container)
	}

	menuTitle := lipgloss.NewStyle().
		Width(20).
		Padding(1, 2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colors.DocumentTextDim).
		Render(m.title)

	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		menuTitle,
		lipgloss.NewStyle().
			Width(26).
			AlignVertical(lipgloss.Center).
			Height(lipgloss.Height(menuTitle)).
			Margin(0, 1).
			Render(m.view.View()),
	)

	helperView := style.HelperView([]style.HelperContent{
		{Key: "↑/↓", Action: "navigate"},
		{Key: "Enter", Action: "select"},
		{Key: "q", Action: "quit"},
	}, lipgloss.Width(mainView), style.NormalView)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		mainView,
		helperView,
	)

	containerWidth := lipgloss.Width(container)
	containerHeight := lipgloss.Height(container)
	containerWidthMargin := (m.width - containerWidth) / 2
	containerHeightMargin := (m.height - containerHeight) / 3
	// containerHeightMargin := 5 // Fixed margin for better visibility

	return lipgloss.NewStyle().Margin(containerHeightMargin, containerWidthMargin, 0).
		Render(container)
}

func menuTitle() string {
	highlight := lipgloss.NewStyle().
		Foreground(colors.DocumentText).
		Bold(true)
	normal := lipgloss.NewStyle().Foreground(colors.DocumentTextDim)

	return lipgloss.JoinVertical(lipgloss.Left,
		highlight.Render("Y")+normal.Render("et"),
		highlight.Render("A")+normal.Render("nother"),
		highlight.Render("Ti")+normal.Render("me"),
		highlight.Render("J")+normal.Render("ournaling"),
		highlight.Render("App")+normal.Render("lication"),
	)
}

func menuView(options []string) components.Radio {
	radio := components.NewRadio(options, components.RadioListView)
	radio.Focus()

	return radio
}
