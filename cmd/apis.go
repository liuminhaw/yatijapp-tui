package main

import (
	"errors"

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

	apiSuccessResponseMsg struct {
		msg      string
		redirect tea.Model
	}
	unexpectedApiResponseMsg error
)

func apiSuccessResponseCmd(msg string, model tea.Model) tea.Cmd {
	return func() tea.Msg {
		return apiSuccessResponseMsg{
			msg:      msg,
			redirect: model,
		}
	}
}

func apiErrorResponseCmd(err error) tea.Msg {
	var le data.LoadApiDataErr
	if errors.As(err, &le) {
		return le
	} else {
		return unexpectedApiResponseMsg(err)
	}
}

func loadTarget(serverURL, uuid, msg string) tea.Msg {
	target, err := data.GetTarget(serverURL, uuid)
	if err != nil {
		return apiErrorResponseCmd(err)
	}

	return getTargetLoadedMsg{
		target: target,
		msg:    msg,
	}
}
