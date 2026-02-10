package main

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/components"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
	"github.com/liuminhaw/yatijapp-tui/pkg/strview"
)

type authView struct {
	name     string
	view     []components.Radio
	page     int
	greeting string

	prev *authView
}

func unauthView() *authView {
	return &authView{
		name: "unauth",
		view: menuView([][]string{
			{"Sign in", "Sign up", "", "", "Forget password"},
		}),
		page: 0,
		prev: nil,
	}
}

func menuAuthView(name string) *authView {
	return &authView{
		name: "menu",
		view: menuView([][]string{
			{"Targets", "Actions", "Sessions", "", "Sign out"},
			{"Preferences"},
		}),
		page:     0,
		greeting: fmt.Sprintf("Welcome, %s", name),
		prev:     nil,
	}
}

func preferencesAuthView(prev *authView) *authView {
	return &authView{
		name: "preferences",
		view: menuView([][]string{
			{"Filter"},
		}),
		page:     0,
		greeting: "Preferences",
		prev:     prev,
	}
}

func filterAuthView(prev *authView) *authView {
	return &authView{
		name: "filter",
		view: menuView([][]string{
			{"Targets", "Actions", "Sessions"},
		}),
		greeting: "Preferences - Filter",
		prev:     prev,
	}
}

type menuPage struct {
	cfg config

	title       string
	view        *authView
	currentView string

	authView   *authView
	unauthView *authView

	width  int
	height int

	spinner spinner.Model
	loading bool

	popupModels []tea.Model
	popup       string

	// greeting string
	msg   string
	error error
}

func newMenuPage(cfg config, width, height int) menuPage {
	page := menuPage{
		cfg:         cfg,
		title:       menuTitle(),
		unauthView:  unauthView(),
		width:       width,
		height:      height,
		spinner:     spinner.New(spinner.WithSpinner(spinner.Meter)),
		loading:     true,
		popupModels: []tea.Model{},
	}
	page.view = page.unauthView

	return page
}

func (m menuPage) loadLoginUser() tea.Cmd {
	return func() tea.Msg {
		user, err := data.GetCurrentUser(m.cfg.apiEndpoint, m.cfg.authClient)
		if err != nil {
			return err
		}

		return apiSuccessResponseMsg{
			msg:      user.Name,
			source:   m,
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
		if m.popup != "" {
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			selected := m.view.view[m.view.page].Selected()
			switch m.view.name {
			case "menu":
				switch selected {
				case "Targets":
					return m, switchToTargetsCmd
				case "Actions":
					return m, switchToActionsCmd
				case "Sessions":
					return m, switchToSessionsCmd
				case "Sign out":
					return m, m.signout()
				case "Preferences":
					// m.cfg.logger.Info("switch to preference", "view", fmt.Sprintf("%+v", m.view))
					m.authView = preferencesAuthView(m.view)
					m.view = m.authView
					// m.cfg.logger.Info("on preference", "prev", fmt.Sprintf("%+v", m.view.prev))
				}
			case "unauth":
				switch selected {
				case "Sign in":
					return m, switchToSigninCmd
				case "Sign up":
					return m, switchToSignupCmd
				case "Forget password":
					return m, switchToResetPasswordCmd
				}
			case "preferences":
				switch selected {
				case "Filter":
					m.authView = filterAuthView(m.view)
					m.view = m.authView
				}
			case "filter":
				switch selected {
				case "Targets":
					return m, switchToFilterCmd(m.cfg.preferences.GetFilter(data.RecordTypeTarget))
				case "Actions":
					return m, switchToFilterCmd(m.cfg.preferences.GetFilter(data.RecordTypeAction))
				case "Sessions":
					return m, switchToFilterCmd(m.cfg.preferences.GetFilter(data.RecordTypeSession))
				}
			}
			return m, nil
		case "ctrl+/", "ctrl+_":
			return m, switchToSearchCmd(data.RecordTypeAll)
		case "<":
			if m.view.prev != nil {
				// m.cfg.logger.Info("switch to previous auth view", "prev", fmt.Sprintf("%+v", m.view.prev))
				m.authView = m.view.prev
				m.view = m.authView
			}
			return m, nil
		case "right", "l":
			if len(m.view.view) > 1 {
				m.view.page = 1
			}
		case "left", "h":
			if len(m.view.view) > 1 {
				m.view.page = 0
			}
		case "up", "down", "k", "j":
			m.msg = ""
		}
	case showSearchMsg:
		popupModel := newSearchPage(
			m.cfg, msg.scope, style.ViewSize{Width: m.width, Height: m.height}, "", m,
		)
		m.popupModels = append(m.popupModels, popupModel)
		m.popup = m.popupModels[len(m.popupModels)-1].View()
	case switchToPreviousMsg:
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.loadLoginUser())
	case apiSuccessResponseMsg:
		if isExactType[menuPage](msg.source) {
			preferences, err := data.GetPreferences(m.cfg.apiEndpoint, m.cfg.authClient)
			if err != nil {
				var ne data.NotFoundApiDataErr
				if errors.As(err, &ne) {
					m.cfg.logger.Info("no user preferences found, load default preferences")
					preferences = data.DefaultPreferences
				} else {
					m.cfg.logger.Error("failed to load user preferences")
					m.error = err
					return m, cmd
				}
			}
			m.cfg.preferences = &preferences
			m.cfg.logger.Info(
				"loaded user preferences",
				slog.Any("preferences", fmt.Sprintf("preferences: %+v", preferences)),
			)

			m.authView = menuAuthView(msg.msg)
			m.view = m.authView
			m.view.page = 0
			m.loading = false
			return m, obtainPreferencesCmd(preferences)
		} else {
			m.cfg.logger.Info("menu page receive api success response from other source")
			m.msg = msg.msg
			return m, m.loadLoginUser()
		}
	case data.UnauthorizedApiDataErr:
		m.cfg.logger.Error(msg.Err.Error(), slog.String("action", "load login user"))
		m.view = m.unauthView
		m.loading = false
	case data.UnexpectedApiDataErr:
		m.cfg.logger.Error(msg.Error(), slog.String("action", "load login user"))
		m.error = errors.New(msg.Msg)
		m.loading = false
	case error:
		m.cfg.logger.Error(msg.Error(), slog.String("action", "load login user"))
		m.error = msg
		m.loading = false
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if len(m.popupModels) > 0 {
		lastIndex := len(m.popupModels) - 1
		m.popupModels[lastIndex], cmd = m.popupModels[lastIndex].Update(msg)
		m.popup = m.popupModels[lastIndex].View()
		return m, cmd
	}

	m.view.view[m.view.page], cmd = m.view.view[m.view.page].Update(msg)

	return m, cmd
}

func (m menuPage) View() string {
	if m.error != nil {
		container := lipgloss.JoinVertical(
			lipgloss.Center,
			style.TitleBarView([]string{"Menu"}, viewWidth, false),
			style.ErrorView(
				style.ViewSize{Width: 80, Height: 16},
				m.error,
				[]style.HelperContent{{Key: "q", Action: "quit"}},
			),
		)

		return style.ContainerStyle(m.width, container, 5).Render(container)
	}

	if m.loading {
		msg := style.Document.NormalDim.Bold(true).Render("Yatijapp")
		m.spinner.Style = style.Document.Highlight
		container := lipgloss.NewStyle().
			Width(50).
			Height(10).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("%s %s", m.spinner.View(), msg))

		return style.ContainerStyle(m.width, container, 5).Render(container)
	}

	greetingView := lipgloss.NewStyle().
		Width(50).
		Foreground(colors.Primary).
		Bold(true).
		Render(m.view.greeting)

	menuTitle := style.BorderStyle["normal"].Width(20).Padding(1, 2).Render(m.title)

	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		menuTitle,
		lipgloss.NewStyle().
			Width(26).
			AlignVertical(lipgloss.Center).
			Height(lipgloss.Height(menuTitle)).
			Margin(0, 1).
			Render(m.view.view[m.view.page].View()),
	)

	msgView := style.MsgStyle.Width(lipgloss.Width(mainView)).
		AlignHorizontal(lipgloss.Center).
		Render(m.msg)

	helper := []style.HelperContent{
		{Key: "↑/↓", Action: "navigate"},
		{Key: "Enter", Action: "select"},
		{Key: "ctrl+/", Action: "search"},
		{Key: "q", Action: "quit"},
	}
	if len(m.view.view) > 1 && m.view.page == 0 {
		helper = append(helper, style.HelperContent{Key: "→", Action: "more"})
	} else if len(m.view.view) > 1 && m.view.page == 1 {
		helper = append(helper, style.HelperContent{Key: "←", Action: "back"})
	}

	if m.view.prev != nil {
		helper = append(helper, style.HelperContent{Key: "<", Action: "back"})
	}

	helperView := style.HelperView(helper, lipgloss.Width(mainView)+20)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		greetingView,
		mainView,
		msgView,
		helperView,
	)

	containerWidth := lipgloss.Width(container)
	containerHeight := lipgloss.Height(container)
	containerWidthMargin := (m.width - containerWidth) / 2
	containerHeightMargin := (m.height - containerHeight) / 3

	container = lipgloss.NewStyle().
		Margin(0, containerWidthMargin, 0).
		Render(container)

	if m.popup != "" {
		overlayX := lipgloss.Width(container)/2 - lipgloss.Width(m.popup)/2
		overlayY := lipgloss.Height(container)/2 - lipgloss.Height(m.popup)/2
		container = strview.PlaceOverlay(overlayX, overlayY, m.popup, container)
	}
	container = lipgloss.NewStyle().
		Margin(containerHeightMargin, 0, 0).
		Render(container)

	return container
}

func (m menuPage) signout() tea.Cmd {
	return func() tea.Msg {
		_, err := data.Signout(m.cfg.apiEndpoint, m.cfg.authClient)
		if err != nil {
			return err
		}

		return apiSuccessResponseMsg{
			msg:      "sign out successfully",
			source:   Water{},
			redirect: m,
		}
	}
}

func menuTitle() string {
	highlight := lipgloss.NewStyle().
		Foreground(colors.Text).
		Bold(true)
	normal := lipgloss.NewStyle().Foreground(colors.TextMuted)

	return lipgloss.JoinVertical(lipgloss.Left,
		highlight.Render("Y")+normal.Render("et"),
		highlight.Render("A")+normal.Render("nother"),
		highlight.Render("Ti")+normal.Render("me"),
		highlight.Render("J")+normal.Render("ournaling"),
		highlight.Render("App")+normal.Render("lication"),
	)
}

func menuView(options [][]string) []components.Radio {
	var radios []components.Radio
	for _, opts := range options {
		radio := components.NewRadio(opts, components.RadioListView, 26)
		radio.Focus()
		radios = append(radios, radio)
	}

	return radios
}
