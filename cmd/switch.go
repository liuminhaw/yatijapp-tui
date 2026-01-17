package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
)

type (
	switchToPreviousMsg struct {
		model tea.Model
	}
	switchToRecordsMsg struct {
		recordType data.RecordType
		sourceUUID string
	}
	switchToViewMsg struct {
		recordType data.RecordType
		uuid       string
	}

	switchToTargetsMsg       struct{}
	switchToActionsMsg       struct{ parents data.RecordParents }
	switchToSessionsMsg      struct{ parents data.RecordParents }
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
		record yatijappRecord
	}
	switchToActionViewMsg struct {
		uuid string
	}
	switchToActionCreateMsg struct {
		parents data.RecordParents
	}
	switchToActionEditMsg struct {
		record yatijappRecord
	}
	switchToTargetSelectorMsg struct{}

	switchToSessionViewMsg struct{ uuid string }
	switchToSessionEditMsg struct {
		record yatijappRecord
	}
	switchToFilterMsg     struct{ f data.RecordFilter }
	switchToSearchListMsg struct{ query string }

	showSearchMsg        struct{ scope data.RecordType }
	showSessionCreateMsg struct {
		parents data.RecordParents
	}
)

var (
	switchToTargetsCmd       = func() tea.Msg { return switchToTargetsMsg{} }
	switchToActionsCmd       = func() tea.Msg { return switchToActionsMsg{} }
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

func switchToRecordsCmd(record yatijappRecord) tea.Cmd {
	return func() tea.Msg {
		switch record.GetActualType() {
		case data.RecordTypeTarget:
			return switchToActionsMsg{
				parents: data.RecordParents{
					data.RecordTypeTarget: {UUID: record.GetUUID(), Title: record.GetTitle()},
				},
			}
		case data.RecordTypeAction:
			return switchToSessionsMsg{
				parents: data.RecordParents{
					data.RecordTypeTarget: {
						UUID:  record.GetParentsUUID()[data.RecordTypeTarget],
						Title: record.GetParentsTitle()[data.RecordTypeTarget],
					},
					data.RecordTypeAction: {UUID: record.GetUUID(), Title: record.GetTitle()},
				},
			}
		}

		panic("unsupported record type in switchToRecordsCmd")
	}
}

func switchToViewCmd(recordType data.RecordType, uuid string) tea.Cmd {
	return func() tea.Msg {
		switch recordType {
		case data.RecordTypeTarget:
			return switchToTargetViewMsg{uuid: uuid}
		case data.RecordTypeAction:
			return switchToActionViewMsg{uuid: uuid}
		case data.RecordTypeSession:
			return switchToSessionViewMsg{uuid: uuid}
		}

		panic("unsupported record type in switchToViewCmd")
	}
}

func switchToCreateCmd(recordType data.RecordType, parents data.RecordParents) tea.Cmd {
	return func() tea.Msg {
		switch recordType {
		case data.RecordTypeTarget:
			return switchToTargetCreateMsg{}
		case data.RecordTypeAction:
			return switchToActionCreateMsg{parents: parents}
		case data.RecordTypeSession:
			return showSessionCreateMsg{parents: parents}
		}

		panic("unsupported record type in switchToCreateCmd")
	}
}

func switchToEditCmd(recordType data.RecordType, record yatijappRecord) tea.Cmd {
	return func() tea.Msg {
		switch recordType {
		case data.RecordTypeTarget:
			return switchToTargetEditMsg{record: record}
		case data.RecordTypeAction:
			return switchToActionEditMsg{record: record}
		case data.RecordTypeSession:
			return switchToSessionEditMsg{record: record}
		}

		panic("unsupported record type in switchToEditCmd")
	}
}

func switchToTargetViewCmd(uuid string) tea.Cmd {
	return func() tea.Msg {
		return switchToTargetViewMsg{uuid: uuid}
	}
}

func switchToActionViewCmd(uuid string) tea.Cmd {
	return func() tea.Msg {
		return switchToActionViewMsg{uuid: uuid}
	}
}

func switchToFilterCmd(f data.RecordFilter) tea.Cmd {
	return func() tea.Msg {
		return switchToFilterMsg{f: f}
	}
}

func switchToSearchCmd(scope data.RecordType) tea.Cmd {
	return func() tea.Msg {
		return showSearchMsg{scope: scope}
	}
}

func switchToSearchListCmd(q string) tea.Cmd {
	return func() tea.Msg {
		return switchToSearchListMsg{query: q}
	}
}

// func sendWindowSizeCmd(width, height int) tea.Cmd {
// 	return func() tea.Msg {
// 		return tea.WindowSizeMsg{
// 			Width:  width,
// 			Height: height,
// 		}
// 	}
// }
