package main

import (
	"errors"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/model"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

const (
	signupFormWidth = 40
)

type signupPage struct {
	cfg    config
	action cmdAction

	fields  []Focusable
	focused int

	width  int
	height int

	err  error
	prev tea.Model
}

func newSignupPage(cfg config, action cmdAction, size style.ViewSize, prev tea.Model) signupPage {
	m := signupPage{
		cfg:    cfg,
		action: action,
		width:  size.Width,
		height: size.Height,
		prev:   prev,
	}

	m.setFields()

	return m
}

func (m *signupPage) setFields() {
	focusables := []Focusable{}

	switch m.action {
	case cmdCreate:
		name := usernameField(signupFormWidth, true)
		email := emailField(signupFormWidth, false)
		password := passwordField(signupFormWidth, false)
		passwordConfirm := passwordConfirmField(
			signupFormWidth,
			false,
			password.(*model.TextInputWrapper),
		)
		focusables = append(focusables, name, email, password, passwordConfirm)
	case cmdUpdate:
		token := tokenField(signupFormWidth, true, "activation token")
		focusables = append(focusables, token)
	}

	m.fields = focusables
	m.focused = 0
	m.err = nil
}

func (m signupPage) Init() tea.Cmd {
	return nil
}

func (m signupPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, switchToPreviousCmd(m.prev)
		case "enter":
			switch m.action {
			case cmdCreate:
				return m, m.signup()
			case cmdUpdate:
				return m, m.activate()
			}
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.fields[m.focused].Validate()
			m.fields[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.fields)
			return m, m.fields[m.focused].Focus()
		case "shift+tab":
			m.fields[m.focused].Validate()
			m.fields[m.focused].Blur()
			m.focused = (m.focused - 1 + len(m.fields)) % len(m.fields)
			return m, m.fields[m.focused].Focus()
		}
	case apiSuccessResponseMsg:
		if m.action == cmdCreate {
			email := m.fields[1]
			password := m.fields[2]
			m.action = cmdUpdate
			m.setFields()
			m.fields = append(m.fields, email, password)
			// m.cfg.logger.Info("Account activated", slog.String("email", email.Value()), slog.String("password", password.Value()))
		}
		return m, nil
	case data.LoadApiDataErr:
		m.cfg.logger.Error(msg.Error(), slog.Int("status", msg.Status), slog.String("action", "user signup"))
		m.err = errors.New(msg.Msg)
	case error:
		m.err = msg
	}

	for i, field := range m.fields {
		retModel, retCmd := field.Update(msg)
		m.fields[i] = retModel.(Focusable)
		if retCmd != nil {
			cmds = append(cmds, retCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m signupPage) View() string {
	var container string

	switch m.action {
	case cmdCreate:
		container = m.signupView()
	case cmdUpdate:
		container = m.activateView()
	}

	return style.ContainerStyle(m.width, container, 5).Render(container)
}

func (m signupPage) signupView() string {
	name := m.fields[0]
	email := m.fields[1]
	password := m.fields[2]
	passwordConfirm := m.fields[3]

	titleView := style.TitleBarView("Sign Up", viewWidth, false)

	form := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Width(signupFormWidth).Margin(1, 0, 0).Render(
			fmt.Sprintf(
				"%s %s\n%s",
				style.FormFieldStyle.Prompt("Username", name.Focused()),
				style.FormFieldStyle.Error.Render(name.Error()),
				style.FormFieldStyle.Content.MarginTop(1).Render(name.View()),
			),
		),
		lipgloss.NewStyle().Width(signupFormWidth).Margin(1, 0, 0).Render(
			fmt.Sprintf(
				"%s %s\n%s",
				style.FormFieldStyle.Prompt("Email", email.Focused()),
				style.FormFieldStyle.Error.Render(email.Error()),
				style.FormFieldStyle.Content.MarginTop(1).Render(email.View()),
			),
		),
		lipgloss.NewStyle().Width(signupFormWidth).Margin(1, 0, 0).Render(
			fmt.Sprintf(
				"%s %s\n%s",
				style.FormFieldStyle.Prompt("Password", password.Focused()),
				style.FormFieldStyle.Error.Render(password.Error()),
				style.FormFieldStyle.Content.
					MarginTop(1).
					Render(password.View()),
			),
		),
		lipgloss.NewStyle().Width(resetPasswordFormWidth).Margin(1, 0, 0).Render(
			fmt.Sprintf(
				"%s %s\n%s\n",
				style.FormFieldStyle.Prompt("Confirm Password", passwordConfirm.Focused()),
				style.FormFieldStyle.Error.Render(passwordConfirm.Error()),
				style.FormFieldStyle.Content.MarginTop(1).Render(passwordConfirm.View()),
			),
		),
	)

	helperContent := []style.HelperContent{
		{Key: "esc", Action: "back"},
		{Key: "tab/shift+tab", Action: "navigate"},
		{Key: "enter", Action: "submit"},
		{Key: "<C-c>", Action: "quit"},
	}
	helperView := style.HelperView(helperContent, viewWidth, style.NormalView)
	msgView := style.ErrorView(style.ViewSize{Width: viewWidth, Height: 1}, m.err, nil)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleView,
		form,
		msgView,
		helperView,
	)
}

func (m signupPage) activateView() string {
	msg := "An email with activation token has been sent to you, please enter the token below to activate your account."
	token := m.fields[0]

	titleView := style.TitleBarView("Activate Account", viewWidth, false)

	form := lipgloss.JoinVertical(
		lipgloss.Center,
		style.MsgStyle.Width(viewWidth).
			AlignHorizontal(lipgloss.Center).
			Margin(1, 0, 0).
			Render(msg),
		lipgloss.NewStyle().Width(signupFormWidth).Margin(1, 0, 0).Render(
			fmt.Sprintf(
				"%s %s\n%s\n",
				style.InputStyle.Selected.Render("Activation Token"),
				style.FormFieldStyle.Error.Render(token.Error()),
				style.FormFieldStyle.Content.MarginTop(1).Render(token.View()),
			),
		),
	)

	helperContent := []style.HelperContent{
		{Key: "esc", Action: "back"},
		{Key: "enter", Action: "submit"},
		{Key: "<C-c>", Action: "quit"},
	}
	helperView := style.HelperView(helperContent, viewWidth, style.NormalView)

	msgView := style.ErrorView(style.ViewSize{Width: viewWidth, Height: 1}, m.err, nil)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleView,
		form,
		msgView,
		helperView,
	)
}

func (m signupPage) validationError() error {
	return fieldsValidation(m.fields, "input validation failed")
}

func (m signupPage) signup() tea.Cmd {
	name := m.fields[0].Value()
	email := m.fields[1].Value()
	password := m.fields[2].Value()

	if err := m.validationError(); err != nil {
		return validationErrorCmd(err)
	}

	request := data.UserRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}
	err := request.Register(m.cfg.serverURL)
	if err != nil {
		return func() tea.Msg { return apiErrorResponseCmd(err) }
	}

	return apiSuccessResponseCmd("", m, m)
}

func (m signupPage) activate() tea.Cmd {
	token := m.fields[0].Value()

	if err := m.validationError(); err != nil {
		return validationErrorCmd(err)
	}

	request := data.UserTokenRequest{Token: token}
	err := request.ActivateUser(m.cfg.serverURL)
	if err != nil {
		return func() tea.Msg { return apiErrorResponseCmd(err) }
	}

	if len(m.fields) != 3 {
		return apiSuccessResponseCmd(
			"Account activated, you can now sign in.",
			m,
			newMenuPage(m.cfg, m.width, m.height),
		)
	}

	email := m.fields[1].Value()
	password := m.fields[2].Value()

	signinReq := data.UserRequest{
		Email:    email,
		Password: password,
	}
	if err := signinReq.Signin(m.cfg.serverURL, m.cfg.authClient.TokenPath); err != nil {
		var le data.LoadApiDataErr
		if errors.As(err, &le) {
			m.cfg.logger.Error(
				le.Error(),
				slog.Int("status", le.Status),
				slog.String("action", "activation auto login"),
			)
		} else {
			m.cfg.logger.Error(err.Error(), slog.String("action", "activation auto login"))
		}

		return apiSuccessResponseCmd(
			"Account activated, you can now sign in.",
			m,
			newMenuPage(m.cfg, m.width, m.height),
		)
	}

	return apiSuccessResponseCmd(
		"Account activated",
		m,
		newMenuPage(m.cfg, m.width, m.height),
	)
}
