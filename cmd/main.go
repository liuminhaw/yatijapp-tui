package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Focusable interface {
	Focus() tea.Cmd
	Focused() bool
	Blur()
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
	Value() string
	SetValue(string) error

	Validate()
	Error() string
}

type config struct {
	serverURL string // http://yatijapp.server.url
	logger    *slog.Logger

	authClient *authclient.AuthClient
}

func main() {
	var cfg config

	flag.StringVar(
		&cfg.serverURL,
		"url",
		"http://localhost:8080",
		"yatijapp server url (https://www.example.com)",
	)
	flag.Parse()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}

	rotater := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/.yatijapp/tui.log", homeDir),
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   false,
	}
	cfg.logger = slog.New(slog.NewJSONHandler(rotater, nil))

	cfg.authClient = &authclient.AuthClient{
		Client:    http.DefaultClient,
		Refresh:   authclient.RefreshToken(cfg.serverURL),
		TokenPath: filepath.Join(homeDir, ".yatijapp", "creds", "token.json"),
	}

	// p := tea.NewProgram(newMainModel(cfg), tea.WithAltScreen(), tea.WithoutCatchPanics())
	p := tea.NewProgram(newMainModel(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

type mainModel struct {
	cfg   config
	ready bool

	// targetSettings targetPage
	active tea.Model
	width  int
	height int
}

func newMainModel(cfg config) mainModel {
	return mainModel{
		cfg: cfg,
	}
}

func (m mainModel) Init() tea.Cmd {
	return nil
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.ready = true
			return m, switchToMenuCmd
		}
	case switchToPreviousMsg:
		m.active = msg.model
	case switchToMenuMsg:
		m.active = newMenuPage(m.cfg, m.width, m.height)
		return m, m.active.Init()
	case switchToSigninMsg:
		m.active = newSigninPage(m.cfg, style.ViewSize{Width: m.width, Height: m.height})
	case switchToSignupMsg:
		m.active = newSignupPage(m.cfg, cmdCreate, style.ViewSize{Width: m.width, Height: m.height}, m.active)
	case switchToResetPasswordMsg:
		m.active = newResetPasswordPage(m.cfg, cmdCreate, style.ViewSize{Width: m.width, Height: m.height}, m.active)
	case switchToTargetsMsg:
		m.active = newTargetListPage(
			m.cfg, style.ViewSize{Width: m.width, Height: m.height}, data.RecordParents{}, m.active,
		)
		return m, m.active.Init()
	case switchToTargetViewMsg:
		m.active = newTargetViewPage2(
			m.cfg,
			msg.uuid,
			style.ViewSize{Width: m.width, Height: m.height},
			style.ViewSize{Width: 80, Height: 20},
			m.active, // Previous model for navigation
		)
		return m, m.active.Init()
	case switchToTargetCreateMsg:
		page, err := newTargetConfigPage(
			m.cfg,
			"New Target",
			style.ViewSize{Width: m.width, Height: m.height},
			nil,
			m.active,
		)
		if err != nil {
			// Switch to error view
			return m, tea.Quit
		}
		m.active = page
	case switchToTargetEditMsg:
		page, err := newTargetConfigPage(
			m.cfg,
			"Modify Target",
			style.ViewSize{Width: m.width, Height: m.height},
			msg.record,
			m.active,
		)
		if err != nil {
			// Switch to error view
			return m, tea.Quit
		}
		m.active = page
	case switchToActionCreateMsg:
		m.cfg.logger.Info("Switching to action create page")
		var record yatijappRecord
		if !msg.parents.IsEmpty() {
			record = data.Action{
				Status:      "queued",
				TargetUUID:  msg.parents.UUID(data.RecordTypeTarget),
				TargetTitle: msg.parents.Title(data.RecordTypeTarget),
			}
		}

		page, err := newActionConfigPage(
			m.cfg,
			"New Action",
			style.ViewSize{Width: m.width, Height: m.height},
			record,
			m.active,
		)
		if err != nil {
			// Switch to error view
			m.cfg.logger.Error(err.Error(), slog.String("action", "switch to action create page"))
			return m, tea.Quit
		}
		m.active = page
	case switchToActionEditMsg:
		page, err := newActionConfigPage(
			m.cfg,
			"Modify Action",
			style.ViewSize{Width: m.width, Height: m.height},
			msg.record,
			m.active,
		)
		if err != nil {
			// Switch to error view
			return m, tea.Quit
		}
		m.active = page
	case switchToActionsMsg:
		m.active = newActionListPage(
			m.cfg,
			style.ViewSize{Width: m.width, Height: m.height},
			msg.parents,
			m.active,
		)
		return m, m.active.Init()
	case switchToActionViewMsg:
		m.active = newActionViewPage(
			m.cfg,
			msg.uuid,
			style.ViewSize{Width: m.width, Height: m.height},
			style.ViewSize{Width: viewWidth, Height: 20},
			m.active, // Previous model for navigation
		)
		return m, m.active.Init()
	case switchToSessionsMsg:
		m.active = newSessionListPage(
			m.cfg,
			style.ViewSize{Width: m.width, Height: m.height},
			msg.parents,
			m.active,
		)
		return m, m.active.Init()
	case switchToSessionViewMsg:
		m.active = newSessionViewPage(
			m.cfg,
			msg.uuid,
			style.ViewSize{Width: m.width, Height: m.height},
			style.ViewSize{Width: viewWidth, Height: 20},
			m.active, // Previous model for navigation
		)
		return m, m.active.Init()
	case switchToTargetSelectorMsg:
		m.active = newTargetSelectorPage(m.cfg, style.ViewSize{Width: m.width, Height: m.height}, m.active)
		return m, m.active.Init()
	case selectorTargetSelectedMsg:
		m.active = msg.model
	case selectorActionSelectedMsg:
		m.active = msg.model
	case apiSuccessResponseMsg:
		m.active = msg.redirect
	}

	if m.active != nil {
		m.active, cmd = m.active.Update(msg)
	}

	return m, cmd
}

func (m mainModel) View() string {
	if m.active == nil {
		return "Loading..."
	}
	return m.active.View()
}
