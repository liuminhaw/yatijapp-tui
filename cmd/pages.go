package main

import (
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
)

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
	internalErrorMsg   struct {
		msg string
		err error
	}

	showSelectorMsg struct {
		selection data.RecordType
	}
	selectorTargetSelectedMsg struct {
		model tea.Model
		title string
		uuid  string
	}
	selectorActionSelectedMsg struct {
		model tea.Model
		title string
		uuid  string
	}
)

var confirmationCmd = func() tea.Msg { return confirmationMsg{} }

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

func selectorSelectedCmd(
	model tea.Model,
	title, uuid string,
	recordType data.RecordType,
) tea.Cmd {
	return func() tea.Msg {
		switch recordType {
		case data.RecordTypeTarget:
			return selectorTargetSelectedMsg{
				model: model,
				title: title,
				uuid:  uuid,
			}
		case data.RecordTypeAction:
			return selectorActionSelectedMsg{
				model: model,
				title: title,
				uuid:  uuid,
			}
		}

		panic("unsupported record type in selectorSelectedCmd")
	}
}

type sourceInfo struct {
	title string
	uuid  string
}

func isExactType[T any](v any) bool {
	_, ok := v.(T)
	return ok
}

type recordsSelection struct {
	records  []yatijappRecord
	selected int
	p        paginator.Model
}

func (rs *recordsSelection) prev() {
	if rs.selected > 0 {
		rs.selected--
		start, _ := rs.p.GetSliceBounds(len(rs.records))
		if rs.selected < start {
			rs.p.PrevPage()
		}
	}
}

func (rs *recordsSelection) next() {
	if rs.selected < len(rs.records)-1 {
		rs.selected++
		_, end := rs.p.GetSliceBounds(len(rs.records))
		if rs.selected >= end {
			rs.p.NextPage()
		}
	}
}

func (rs *recordsSelection) nextPage() {
	if rs.p.OnLastPage() {
		return
	}
	rs.p.NextPage()
	rs.selected, _ = rs.p.GetSliceBounds(len(rs.records))
}

func (rs *recordsSelection) prevPage() {
	if rs.p.OnFirstPage() {
		return
	}
	rs.p.PrevPage()
	rs.selected, _ = rs.p.GetSliceBounds(len(rs.records))
}
