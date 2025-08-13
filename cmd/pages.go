package main

import tea "github.com/charmbracelet/bubbletea"

const (
	viewWidth = 80
	formWidth = 70
)

type cmdAction int

const (
	cmdCreate cmdAction = iota
	cmdUpdate
)

type (
	confirmationMsg struct{}

	validationErrorMsg error
	internalErrorMsg struct {
		msg string
		err error
	}
)

var (
	confirmationCmd = func() tea.Msg { return confirmationMsg{} }
)

func (m internalErrorMsg) Error() string {
	return m.msg + ": " + m.err.Error()
}

func validationErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return validationErrorMsg(err)
	}
}

func internalErrorCmd(msg string, err error) tea.Cmd {
	return func() tea.Msg {
		return internalErrorMsg{
			msg: msg,
			err: err,
		}
	}
}
