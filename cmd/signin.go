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
	signinFormWidth = 40
)

type signinPage struct {
	cfg config
	// mode style.ViewMode

	fields  []Focusable
	focused int

	width  int
	height int

	err error
}

func newSigninPage(cfg config, size style.ViewSize) signinPage {
	focusables := []Focusable{}

	email := emailField(signinFormWidth, true)
	password := passwordField(signinFormWidth, false)

	focusables = append(focusables, email, password)

	return signinPage{
		cfg:    cfg,
		fields: focusables,
		width:  size.Width,
		height: size.Height,
	}
}

func (m signinPage) Init() tea.Cmd {
	return nil
}

func (m signinPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
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
		case "esc":
			return m, switchToMenuCmd
		case "enter":
			return m, m.signin()
		case "ctrl+c":
			return m, tea.Quit
		}
	case data.LoadApiDataErr:
		m.cfg.logger.Error(msg.Error(), slog.Int("status", msg.Status), slog.String("action", "user signin"))
		m.err = errors.New(msg.Msg)
		password := m.fields[1]
		if pType := password.(*model.TextInputWrapper); pType != nil {
			pType.Clear()
		}
	// unexpectedApiResponseMsg, validationErrorMsg
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

func (m signinPage) View() string {
	email := m.fields[0]
	password := m.fields[1]

	titleView := style.TitleBarView("Sign In", viewWidth, false)

	signinForm := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Width(signinFormWidth).Margin(1, 0, 0).Render(
			fmt.Sprintf(
				// "%s\n%s\n%s",
				"%s %s\n%s",
				style.FormFieldStyle.Prompt("User", email.Focused()),
				style.FormFieldStyle.Error.Render(email.Error()),
				style.FormFieldStyle.Content.
					MarginTop(1).
					Render(email.View()),
				// style.FormFieldStyle.Error.Render(email.Error()),
			),
		),
		lipgloss.NewStyle().Width(signinFormWidth).Margin(1, 0, 0).Render(
			fmt.Sprintf(
				"%s %s\n%s\n",
				style.FormFieldStyle.Prompt("Password", password.Focused()),
				style.FormFieldStyle.Error.Render(password.Error()),
				style.FormFieldStyle.Content.
					MarginTop(1).
					Render(password.View()),
			),
		),
	)

	helperContent := []style.HelperContent{
		{Key: "esc", Action: "back"},
		{Key: "enter", Action: "submit"},
		{Key: "tab", Action: "navigate"},
		{Key: "<C-c>", Action: "quit"},
	}
	helperView := style.HelperView(helperContent, viewWidth, style.NormalView)

	msgView := style.ErrorView(style.ViewSize{Width: viewWidth, Height: 1}, m.err, nil)

	container := lipgloss.JoinVertical(
		lipgloss.Center,
		titleView,
		signinForm,
		msgView,
		helperView,
	)

	return style.ContainerStyle(m.width, container, 5).Render(container)
}

func (m signinPage) validationError() error {
	return fieldsValidation(m.fields, "input validation failed")
}

func (m signinPage) signin() tea.Cmd {
	email := m.fields[0].Value()
	password := m.fields[1].Value()

	if err := m.validationError(); err != nil {
		return validationErrorCmd(err)
	}

	request := data.UserRequest{
		Email:    email,
		Password: password,
	}
	if err := request.Signin(m.cfg.serverURL, m.cfg.authClient.TokenPath); err != nil {
		return func() tea.Msg { return apiErrorResponseCmd(err) }
	}

	return apiSuccessResponseCmd(
		"Signed in successfully",
		m,
		newMenuPage(m.cfg, m.width, m.height),
	)
}
