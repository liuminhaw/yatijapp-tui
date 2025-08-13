package main

import (
	"flag"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type Focusable interface {
	Focus() tea.Cmd
	Focused() bool
	Blur()
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
	Value() string

	Validate()
	Error() string
}

type termSize struct {
	width  int
	height int
}

type config struct {
	serverURL string // http://yatijapp.server.url
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

	p := tea.NewProgram(newMainModel(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

type mainModel struct {
	cfg config

	menu menuPage
	// targets        targetListPage
	targetSettings targetPage
	active         tea.Model
	width          int
	height         int
}

func newMainModel(cfg config) mainModel {
	// target := newTargetPage()
	menu := newMenuPage()

	return mainModel{
		cfg:  cfg,
		menu: menu,
		// targetSettings: target,
		active: menu,
	}
}

func (m mainModel) Init() tea.Cmd {
	return m.active.Init()
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case switchToPreviousMsg:
		m.active = msg.model
	case switchToMenuMsg:
		m.active = m.menu
		return m, sendWindowSizeCmd(m.width, m.height)
	case switchToTargetsMsg:
		m.active = newTargetListPage(m.cfg, m.width, m.height)
		return m, m.active.Init()
	case switchToTargetViewMsg:
		m.active = newTargetViewPage(
			m.cfg,
			msg.uuid,
			style.ViewSize{Width: m.width, Height: m.height},
			style.ViewSize{Width: 80, Height: 20},
			m.active, // Previous model for navigation
		)
		return m, m.active.Init()
	case switchToTargetCreateMsg:
		page, err := newTargetPage(
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
		page, err := newTargetPage(
			m.cfg,
			"Modify Target",
			style.ViewSize{Width: m.width, Height: m.height},
			&msg.data,
			m.active,
		)
		if err != nil {
			// Switch to error view
			return m, tea.Quit
		}
		m.active = page
	case apiSuccessResponseMsg:
		m.active = msg.redirect
	}

	m.active, cmd = m.active.Update(msg)

	return m, cmd
}

func (m mainModel) View() string {
	return m.active.View()
}

// type viewField struct {
// 	prompt  string
// 	content string
// 	focused bool
// }
//
// // func newViewField(prompt, content string, focused bool) viewField {
// func newViewField(prompt string, field Focusable) viewField {
// 	// field := viewField{
// 	// 	prompt:  prompt,
// 	// 	content: content,
// 	// 	focused: focused,
// 	// }
//
// 	return viewField{
// 		prompt:  prompt,
// 		content: field.View(),
// 		focused: field.Focused(),
// 	}
// }
//
// func (f viewField) contentStyle() lipgloss.Style {
// 	return style.InputStyle.Document
// }
//
// func (f viewField) promptStyle() lipgloss.Style {
// 	if f.focused {
// 		return style.InputStyle.Selected
// 	}
// 	return style.InputStyle.Prompt
// }
//
// func (f viewField) helperStyle() lipgloss.Style {
// 	return style.InputStyle.Helper
// }
