package main

import (
	"log/slog"
	"slices"
	"strconv"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
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

func (rs *recordsSelection) setRecords(msg allRecordsLoadedMsg, logger *slog.Logger) {
	rs.metadata = msg.metadata
	rs.records = msg.records

	rs.p.SetTotalPages(len(rs.records))
	totalPage := rs.p.TotalPages
	start, end := rs.p.GetSliceBounds(len(rs.records))

	if slices.Contains(msg.events, "prev") {
		rs.p.Page = rs.p.TotalPages - 1
		rs.selected = end - rs.p.PerPage
	}
	if slices.Contains(msg.events, "next") {
		rs.p.Page = 0
		rs.selected = start
	}

	if msg.metadata.CurrentPage == msg.metadata.FirstPage && rs.p.Page == 0 {
		if rs.selected > rs.offset {
			rs.selected -= rs.offset
		} else if rs.selected == rs.offset {
			rs.selected--
		}
	}
	if msg.metadata.CurrentPage != msg.metadata.FirstPage {
		rs.records = append(make([]yatijappRecord, rs.offset), rs.records...)
		rs.p.Page++
		rs.p.TotalPages++
		if rs.selected < rs.offset {
			rs.selected += rs.offset
		}
	}

	if msg.metadata.CurrentPage != msg.metadata.LastPage {
		rs.p.TotalPages++
	}
	logger.Info(
		"Records page",
		slog.Int("page", rs.p.Page),
		slog.Int("total_pages", rs.p.TotalPages),
	)

	if rs.p.Page > totalPage-1 {
		rs.p.Page = totalPage - 1
	}
	if rs.selected > end && rs.selected > 0 {
		rs.selected = end - 1
	} else if rs.selected == end && rs.selected > 0 {
		rs.selected = start
	}
}

func (rs *recordsSelection) prev() tea.Cmd {
	// logger.Info("Calling prev", slog.String("rs metadata", fmt.Sprintf("%+v", rs.metadata)))
	// logger.Info("Calling prev", slog.String("rs paginator", fmt.Sprintf("%+v", rs.p)))
	// logger.Info("Calling prev", slog.Int("rs selected", rs.selected))
	if rs.selected > 0 {
		if rs.metadata.CurrentPage != rs.metadata.FirstPage && rs.selected == rs.offset {
			// logger.Info("Loading more records for prev", slog.String("action", "prev"))
			rs.query["page"] = strconv.Itoa(rs.metadata.CurrentPage - 1)
			// logger.Info("Call load more records cmd")
			return loadMoreRecords("prev")
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

	// logger.Info("Next", slog.Int("end", end))
	if rs.metadata.CurrentPage != rs.metadata.LastPage && rs.p.Page == rs.p.TotalPages-2 {
		// logger.Info("Loading more records for next", slog.String("action", "next"))
		rs.query["page"] = strconv.Itoa(rs.metadata.CurrentPage + 1)
		return loadMoreRecords("next")
	}

	return nil
}

func (rs *recordsSelection) nextPage() tea.Cmd {
	// logger.Info("Calling next page", slog.String("rs metadata", fmt.Sprintf("%+v", rs.metadata)))
	// logger.Info("Calling next page", slog.String("rs paginator", fmt.Sprintf("%+v", rs.p)))
	if rs.p.OnLastPage() {
		// logger.Info("Already on last page", slog.String("action", "next page"))
		return nil
	}
	if rs.metadata.CurrentPage != rs.metadata.LastPage && rs.p.Page == rs.p.TotalPages-2 {
		// logger.Info("Loading more records for next page", slog.String("action", "next page"))
		rs.query["page"] = strconv.Itoa(rs.metadata.CurrentPage + 1)
		return loadMoreRecords("next")
	}
	rs.p.NextPage()
	rs.selected, _ = rs.p.GetSliceBounds(len(rs.records))

	return nil
}

func (rs *recordsSelection) prevPage() tea.Cmd {
	// logger.Info("Calling prev page", slog.String("rs metadata", fmt.Sprintf("%+v", rs.metadata)))
	// logger.Info("Calling prev page", slog.String("rs paginator", fmt.Sprintf("%+v", rs.p)))
	// logger.Info("Calling prev page", slog.Int("rs selected", rs.selected))
	if rs.p.OnFirstPage() {
		return nil
	}
	if rs.metadata.CurrentPage != rs.metadata.FirstPage && rs.p.Page == 1 {
		// logger.Info("Loading more records for prev page", slog.String("action", "prev page"))
		rs.query["page"] = strconv.Itoa(rs.metadata.CurrentPage - 1)
		return loadMoreRecords("prev")
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
