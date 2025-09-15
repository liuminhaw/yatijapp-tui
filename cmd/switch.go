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
	switchToActivitiesMsg    struct{ info sourceInfo }
	switchToSessionsMsg      struct{ activityUUID string }
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
	switchToActivityViewMsg struct {
		uuid string
	}
	switchToActivityCreateMsg struct {
		parentTitle string
		parentUUID  string
	}
	switchToActivityEditMsg struct {
		record yatijappRecord
	}
	switchToTargetSelectorMsg struct{}
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

func switchToRecordsCmd(recordType data.RecordType, srcUUID, srcTitle string) tea.Cmd {
	return func() tea.Msg {
		switch recordType {
		case data.RecordTypeTarget:
			return switchToActivitiesMsg{info: sourceInfo{uuid: srcUUID, title: srcTitle}}
		case data.RecordTypeActivity:
			return switchToSessionsMsg{}
		}

		panic("unsupported record type in switchToRecordsCmd")
	}
}

func switchToViewCmd(recordType data.RecordType, uuid string) tea.Cmd {
	return func() tea.Msg {
		switch recordType {
		case data.RecordTypeTarget:
			return switchToTargetViewMsg{uuid: uuid}
		case data.RecordTypeActivity:
			return switchToActivityViewMsg{uuid: uuid}
		}

		panic("unsupported record type in switchToViewCmd")
	}
}

func switchToCreateCmd(recordType data.RecordType, parentUUID, parentTitle string) tea.Cmd {
	return func() tea.Msg {
		switch recordType {
		case data.RecordTypeTarget:
			return switchToTargetCreateMsg{}
		case data.RecordTypeActivity:
			return switchToActivityCreateMsg{parentUUID: parentUUID, parentTitle: parentTitle}
		}

		panic("unsupported record type in switchToCreateCmd")
	}
}

func switchToEditCmd(recordType data.RecordType, record yatijappRecord) tea.Cmd {
	return func() tea.Msg {
		switch recordType {
		case data.RecordTypeTarget:
			return switchToTargetEditMsg{record: record}
		case data.RecordTypeActivity:
			return switchToActivityEditMsg{record: record}
		}

		panic("unsupported record type in switchToEditCmd")
	}
}

func switchToTargetViewCmd(uuid string) tea.Cmd {
	return func() tea.Msg {
		return switchToTargetViewMsg{uuid: uuid}
	}
}

func switchToActivityViewCmd(uuid string) tea.Cmd {
	return func() tea.Msg {
		return switchToActivityViewMsg{uuid: uuid}
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
