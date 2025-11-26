package main

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
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

func main() {
	vConf := viper.New()

	flag.String("api-endpoint", "https://api.yatij.app", "yatijapp server api endpoint")
	flag.String("display-mode", "auto", "display mode: light | dark | auto")
	flag.Parse()

	cfg, err := configSetup(vConf)
	if err != nil {
		panic(err)
	}

	switch cfg.displayMode {
	case "light":
		cfg.logger.Info("Using light display mode")
		lipgloss.SetHasDarkBackground(false)
	case "dark":
		cfg.logger.Info("Using dark display mode")
		lipgloss.SetHasDarkBackground(true)
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
			style.ViewSize{Width: viewWidth, Height: 20},
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
	case switchToSessionEditMsg:
		page, err := newSessionConfigPage(
			m.cfg,
			"Modify Session",
			style.ViewSize{Width: m.width, Height: m.height},
			msg.record,
			m.active,
		)
		if err != nil {
			// TODO: Switch to error view
			return m, tea.Quit
		}
		m.active = page
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
