package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
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
		source   tea.Model
		redirect tea.Model
	}
)

func loadTarget(serverURL, uuid, msg string, client *authclient.AuthClient) tea.Msg {
	target, err := data.GetTarget(serverURL, uuid, client)
	if err != nil {
		return err
	}

	return getTargetLoadedMsg{
		target: target,
		msg:    msg,
	}
}
