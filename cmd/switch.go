package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
)

type (
	switchToPreviousMsg struct {
		model tea.Model
	}

	switchToTargetsMsg       struct{}
	switchToActivitiesMsg    struct{}
	switchToSessionsMsg      struct{}
	switchToMenuMsg          struct{}
	switchToSigninMsg        struct{}
	switchToSignupMsg        struct{}
	switchToResetPasswordMsg struct{}
	switchToHelperListMsg    struct{}
	switchToTargetViewMsg    struct {
		uuid string
	}
	switchToTargetCreateMsg struct{}
	switchToTargetEditMsg   struct {
		data data.Target
	}
)

var (
	switchToTargetsCmd       = func() tea.Msg { return switchToTargetsMsg{} }
	switchToActivitiesCmd    = func() tea.Msg { return switchToActivitiesMsg{} }
	switchToSessionsCmd      = func() tea.Msg { return switchToSessionsMsg{} }
	switchToMenuCmd          = func() tea.Msg { return switchToMenuMsg{} }
	switchToSigninCmd        = func() tea.Msg { return switchToSigninMsg{} }
	switchToSignupCmd        = func() tea.Msg { return switchToSignupMsg{} }
	switchToResetPasswordCmd = func() tea.Msg { return switchToResetPasswordMsg{} }
	switchToTargetCreateCmd  = func() tea.Msg { return switchToTargetCreateMsg{} }
	switchToHelperListCmd    = func() tea.Msg { return switchToHelperListMsg{} }
)

func switchToPreviousCmd(model tea.Model) tea.Cmd {
	return func() tea.Msg {
		return switchToPreviousMsg{model: model}
	}
}

func switchToTargetViewCmd(uuid string) tea.Cmd {
	return func() tea.Msg {
		return switchToTargetViewMsg{uuid: uuid}
	}
}

func switchToTargetEditCmd(target data.Target) tea.Cmd {
	return func() tea.Msg {
		return switchToTargetEditMsg{data: target}
	}
}

func sendWindowSizeCmd(width, height int) tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{
			Width:  width,
			Height: height,
		}
	}
}
