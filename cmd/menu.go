package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/components"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type menuPage struct {
	title string
	views components.Radio

	width  int
	height int
}

func newMenuPage() menuPage {
	viewOptions := components.NewRadio(
		[]string{"Targets", "Activities", "Sessions"},
		components.RadioListView,
	)
	viewOptions.Focus() // Set focus on the menu view

	return menuPage{
		title: menuTitle(),
		views: viewOptions,
	}
}

func (m menuPage) Init() tea.Cmd {
	return nil
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
			switch m.views.Selected() {
			case "Targets":
				return m, switchToTargetsCmd
			case "Activities":
				return m, switchToActivitiesCmd
			case "Sessions":
				return m, switchToSessionsCmd
			}
		}
	}

	m.views, cmd = m.views.Update(msg)

	return m, cmd
}

func (m menuPage) View() string {
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
			Render(m.views.View()),
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
