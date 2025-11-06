package main

import (
	"log/slog"
	"slices"
	"strconv"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
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

	cancelPopupMsg struct{}
)

var (
	confirmationCmd = func() tea.Msg { return confirmationMsg{} }
	cancelPopupCmd  = func() tea.Msg { return cancelPopupMsg{} }
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
	p        paginator.Model
	selected int
	offset   int
	query    map[string]string
	metadata data.Metadata
}

func newRecordsSelection(pageSize int) recordsSelection {
	return recordsSelection{
		records: []yatijappRecord{},
		p:       style.NewPaginator(pageSize),
		offset:  pageSize,
		query:   make(map[string]string),
		// query: map[string]string{"page_size": "5"},
	}
}

func (rs *recordsSelection) view() string {
	prevArrow := lipgloss.NewStyle().Foreground(colors.HelperTextDim).Render("◂")
	nextArrow := lipgloss.NewStyle().Foreground(colors.HelperTextDim).Render("▸")
	if rs.metadata.CurrentPage != rs.metadata.FirstPage {
		prevArrow = lipgloss.NewStyle().Foreground(colors.Text).Render("◂")
	}
	if rs.metadata.CurrentPage != rs.metadata.LastPage {
		nextArrow = lipgloss.NewStyle().Foreground(colors.Text).Render("▸")
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		prevArrow,
		rs.p.View(),
		nextArrow,
	)
}

func (rs *recordsSelection) setRecords(msg allRecordsLoadedMsg, logger *slog.Logger) {
	rs.metadata = msg.metadata
	rs.records = msg.records

	rs.p.SetTotalPages(len(rs.records))

	var start, end int
	if slices.Contains(msg.events, "prev") {
		rs.p.Page = rs.p.TotalPages - 1
		start, end = rs.p.GetSliceBounds(len(rs.records))
		rs.selected = end - 1
	} else if slices.Contains(msg.events, "prevp") {
		rs.p.Page = rs.p.TotalPages - 1
		start, end = rs.p.GetSliceBounds(len(rs.records))
		logger.Info("Start", slog.Int("start", start))
		logger.Info("End", slog.Int("end", end))
		rs.selected = start
	} else {
		rs.p.Page = 0
		start, end = rs.p.GetSliceBounds(len(rs.records))
		rs.selected = start
	}
}

func (rs *recordsSelection) prev() tea.Cmd {
	// logger.Info("Calling prev", slog.String("rs metadata", fmt.Sprintf("%+v", rs.metadata)))
	// logger.Info("Calling prev", slog.String("rs paginator", fmt.Sprintf("%+v", rs.p)))
	// logger.Info("Calling prev", slog.Int("rs selected", rs.selected))
	if rs.selected >= 0 {
		if rs.metadata.CurrentPage != rs.metadata.FirstPage && rs.selected == 0 {
			rs.query["page"] = strconv.Itoa(rs.metadata.CurrentPage - 1)
			return loadMoreRecords("prev")
		} else if rs.selected == 0 {
			return nil
		}

		rs.selected--
		start, _ := rs.p.GetSliceBounds(len(rs.records))
		if rs.selected < start {
			rs.p.PrevPage()
		}
	}

	return nil
}

func (rs *recordsSelection) next() tea.Cmd {
	// logger.Info("Calling next", slog.String("rs metadata", fmt.Sprintf("%+v", rs.metadata)))
	// logger.Info("Calling next", slog.String("rs paginator", fmt.Sprintf("%+v", rs.p)))
	// logger.Info("Calling next", slog.Int("rs selected", rs.selected))
	if rs.selected < len(rs.records)-1 {
		rs.selected++
		_, end := rs.p.GetSliceBounds(len(rs.records))
		if rs.selected >= end {
			rs.p.NextPage()
		}

		return nil
	}

	if rs.metadata.CurrentPage != rs.metadata.LastPage && rs.p.Page == rs.p.TotalPages-1 {
		rs.query["page"] = strconv.Itoa(rs.metadata.CurrentPage + 1)
		return loadMoreRecords("next")
	}

	return nil
}

func (rs *recordsSelection) nextPage() tea.Cmd {
	// logger.Info("Calling next page", slog.String("rs metadata", fmt.Sprintf("%+v", rs.metadata)))
	// logger.Info("Calling next page", slog.String("rs paginator", fmt.Sprintf("%+v", rs.p)))
	if rs.metadata.CurrentPage != rs.metadata.LastPage && rs.p.Page == rs.p.TotalPages-1 {
		rs.query["page"] = strconv.Itoa(rs.metadata.CurrentPage + 1)
		return loadMoreRecords("nextp")
	}
	if rs.p.OnLastPage() {
		return nil
	}
	rs.p.NextPage()
	rs.selected, _ = rs.p.GetSliceBounds(len(rs.records))

	return nil
}

func (rs *recordsSelection) prevPage() tea.Cmd {
	// logger.Info("Calling prev page", slog.String("rs metadata", fmt.Sprintf("%+v", rs.metadata)))
	// logger.Info("Calling prev page", slog.String("rs paginator", fmt.Sprintf("%+v", rs.p)))
	// logger.Info("Calling prev page", slog.Int("rs selected", rs.selected))
	if rs.metadata.CurrentPage != rs.metadata.FirstPage && rs.p.Page == 0 {
		// logger.Info("Loading more records for prev page", slog.String("action", "prev page"))
		rs.query["page"] = strconv.Itoa(rs.metadata.CurrentPage - 1)
		return loadMoreRecords("prevp")
	}
	if rs.p.OnFirstPage() {
		return nil
	}
	rs.p.PrevPage()
	rs.selected, _ = rs.p.GetSliceBounds(len(rs.records))

	return nil
}

func (rs *recordsSelection) current() yatijappRecord {
	if len(rs.records) == 0 {
		return nil
	}
	return rs.records[rs.selected]
}

func (rs *recordsSelection) hasRecords() bool {
	return len(rs.records) > 0
}
