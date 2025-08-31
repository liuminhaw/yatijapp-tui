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
	resetPasswordFormWidth = 40
)

type resetPasswordPage struct {
	cfg    config
	action cmdAction

	fields  []Focusable
	focused int

	width  int
	height int

	// msg  string
	err  error
	prev tea.Model
}

func newResetPasswordPage(
	cfg config,
	action cmdAction,
	size style.ViewSize,
	prev tea.Model,
) resetPasswordPage {
	m := resetPasswordPage{
		cfg:    cfg,
		action: action,
		width:  size.Width,
		height: size.Height,
		prev:   prev,
	}

	m.setFields()

	return m
}

func (m *resetPasswordPage) setFields() {
	focusables := []Focusable{}

	switch m.action {
	case cmdCreate:
		email := emailField(resetPasswordFormWidth, true)
		focusables = append(focusables, email)
	case cmdUpdate:
		token := tokenField(resetPasswordFormWidth, true, "reset token")
		password := passwordField(resetPasswordFormWidth, false)
		passwordConfirm := passwordConfirmField(
			resetPasswordFormWidth,
			false,
			password.(*model.TextInputWrapper),
		)

		focusables = append(
			focusables,
			token,
			password,
			passwordConfirm,
		)
	}

	m.fields = focusables
	m.focused = 0
	m.err = nil
}

func (m resetPasswordPage) Init() tea.Cmd {
	return nil
}

func (m resetPasswordPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// return m, switchToMenuCmd
			return m, switchToPreviousCmd(m.prev)
		case "enter":
			switch m.action {
			case cmdCreate:
				return m, m.sendResetToken()
			case cmdUpdate:
				return m, m.resetPassword()
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
		// m.msg = msg.msg
		if m.action == cmdCreate {
			m.action = cmdUpdate
			m.setFields()
		}
		return m, nil
	case data.LoadApiDataErr:
		m.cfg.logger.Error(msg.Error(), slog.Int("status", msg.Status), slog.String("action", "reset password"))
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

func (m resetPasswordPage) View() string {
	var container string
	switch m.action {
	case cmdCreate:
		container = m.sendTokenView()
	case cmdUpdate:
		container = m.resetPasswordView()
	}

	return style.ContainerStyle(m.width, container, 5).Render(container)
}

func (m resetPasswordPage) sendTokenView() string {
	email := m.fields[0]

	titleView := style.TitleBarView("Reset Password", viewWidth, false)

	form := lipgloss.NewStyle().Width(resetPasswordFormWidth).Margin(1, 0, 0).Render(
		fmt.Sprintf(
			"%s %s\n%s\n",
			style.InputStyle.Selected.Render("Email"),
			style.FormFieldStyle.Error.Render(email.Error()),
			style.FormFieldStyle.Content.MarginTop(1).Render(email.View()),
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

func (m resetPasswordPage) resetPasswordView() string {
	msg := "An email with reset token has been sent to you, please enter the token and new password below to reset your password."
	token := m.fields[0]
	password := m.fields[1]
	passwordConfirm := m.fields[2]

	titleView := style.TitleBarView("Reset Password", viewWidth, false)

	form := lipgloss.JoinVertical(
		lipgloss.Center,
		style.MsgStyle.Width(viewWidth).
			AlignHorizontal(lipgloss.Center).
			Margin(1, 0, 0).
			Render(msg),
		lipgloss.NewStyle().Width(resetPasswordFormWidth).Margin(1, 0, 0).Render(
			fmt.Sprintf(
				"%s %s\n%s",
				style.FormFieldStyle.Prompt("Reset Token", token.Focused()),
				style.FormFieldStyle.Error.Render(token.Error()),
				style.FormFieldStyle.Content.MarginTop(1).Render(token.View()),
			),
		),
		lipgloss.NewStyle().Width(resetPasswordFormWidth).Margin(1, 0, 0).Render(
			fmt.Sprintf(
				"%s %s\n%s",
				style.FormFieldStyle.Prompt("New Password", password.Focused()),
				style.FormFieldStyle.Error.Render(password.Error()),
				style.FormFieldStyle.Content.MarginTop(1).Render(password.View()),
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

func (m resetPasswordPage) validationError() error {
	return fieldsValidation(m.fields, "input validation failed")
}

func (m resetPasswordPage) sendResetToken() tea.Cmd {
	email := m.fields[0].Value()

	if err := m.validationError(); err != nil {
		return validationErrorCmd(err)
	}

	request := data.ResetPasswordTokenRequest{
		Email: email,
	}
	message, err := request.Do(m.cfg.serverURL)
	if err != nil {
		return func() tea.Msg { return apiErrorResponseCmd(err) }
	}

	return apiSuccessResponseCmd(message.Message, m, m)
}

func (m resetPasswordPage) resetPassword() tea.Cmd {
	token := m.fields[0].Value()
	pasword := m.fields[1].Value()

	if err := m.validationError(); err != nil {
		return validationErrorCmd(err)
	}

	request := data.UserTokenRequest{
		Token:    token,
		Password: pasword,
	}
	message, err := request.ResetPassword(m.cfg.serverURL)
	if err != nil {
		return func() tea.Msg { return apiErrorResponseCmd(err) }
	}

	return apiSuccessResponseCmd(message.Message, m, newMenuPage(m.cfg, m.width, m.height))
}
