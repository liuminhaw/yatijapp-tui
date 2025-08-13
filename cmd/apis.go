package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
)

type (
	targetListLoadedMsg struct {
		targets []data.Target
		msg     string
	}
	getTargetLoadedMsg struct {
		target data.Target
		msg    string
	}
	targetDeletedMsg string

	apiResponseErrorMsg    error
	apiSuccessResponseMsg struct {
		msg      string
		redirect tea.Model
	}
)

func apiSuccessResponseCmd(msg string, model tea.Model) tea.Cmd {
	return func() tea.Msg {
		return apiSuccessResponseMsg{
			msg:      msg,
			redirect: model,
		}
	}
}

func loadTarget(serverURL, uuid, msg string) tea.Msg {
	target, err := data.GetTarget(serverURL, uuid)
	if err != nil {
		return apiResponseErrorMsg(err)
	}

	return getTargetLoadedMsg{
		target: target,
		msg:    msg,
	}
}
