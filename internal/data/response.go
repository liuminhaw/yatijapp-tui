package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type respError map[string]string

func (r *respError) UnmarshalJSON(data []byte) error {
	var singleError string
	if err := json.Unmarshal(data, &singleError); err == nil {
		*r = map[string]string{"": singleError}
		return nil
	}
	return json.Unmarshal(data, (*map[string]string)(r))
}

type ErrorResponse struct {
	Err respError `json:"error,omitempty"`
}

func (e ErrorResponse) Error() string {
	if len(e.Err) == 0 {
		return "empty error"
	}
	if len(e.Err) == 1 {
		if v, ok := e.Err[""]; ok {
			return v
		}
		for k, v := range e.Err {
			return fmt.Sprintf("%s - %s", k, v)
		}
	}

	var buf strings.Builder
	for k, v := range e.Err {
		if k == "" {
			buf.WriteString(fmt.Sprintf("%s, ", v))
		} else {
			buf.WriteString(fmt.Sprintf("%s: %s, ", k, v))
		}
	}
	return strings.TrimRight(strings.TrimSpace(buf.String()), ",")
}

type Message struct {
	Message string `json:"message"`
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitzero"`
	PageSize     int `json:"page_size,omitzero"`
	FirstPage    int `json:"first_page,omitzero"`
	LastPage     int `json:"last_page,omitzero"`
	TotalRecords int `json:"total_records,omitzero"`
}

type RecordType string

func (rt RecordType) GetParentType() (RecordType, bool) {
	switch rt {
	case RecordTypeAction:
		return RecordTypeTarget, true
	case RecordTypeSession:
		return RecordTypeAction, true
	default:
		return "", false
	}
}

const (
	RecordTypeTarget  RecordType = "Target"
	RecordTypeAction  RecordType = "Action"
	RecordTypeSession RecordType = "Session"
)

type Target struct {
	UUID         string       `json:"uuid"`
	CreatedAt    time.Time    `json:"created_at"`
	DueDate      sql.NullTime `json:"due_date"`
	UpdatedAt    time.Time    `json:"updated_at"`
	LastActive   time.Time    `json:"last_active"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Notes        string       `json:"notes"`
	Status       string       `json:"status"`
	Version      int32        `json:"version"`
	HasNotes     bool         `json:"has_notes"`
	ActionsCount int64        `json:"actions_count"`
}

func (t Target) ListItemView(hasSrc, chosen bool, width int) string {
	return listPageItemView(listItemData{
		title:  t.Title,
		status: t.Status,
	}, chosen, width)
}

func (t Target) ListItemDetailView(hasSrc bool, width int) string {
	var due *time.Time
	if t.DueDate.Valid {
		due = &t.DueDate.Time
	}
	return listPageItemDetail(listItemData{
		title:         t.Title,
		description:   t.Description,
		status:        t.Status,
		due:           due,
		itemType:      RecordTypeTarget,
		hasNotes:      t.HasNotes,
		childrenCount: t.ActionsCount,
	}, width)
}

func (t Target) GetActualType() RecordType { return RecordTypeTarget }
func (t Target) GetUUID() string           { return t.UUID }
func (t Target) GetTitle() string          { return t.Title }
func (t Target) GetDescription() string    { return t.Description }
func (t Target) GetStatus() string         { return t.Status }
func (t Target) GetNote() string           { return t.Notes }
func (t Target) GetDueDate() (time.Time, bool) {
	if t.DueDate.Valid {
		return t.DueDate.Time, true
	}
	return time.Time{}, false
}
func (t Target) GetCreatedAt() time.Time  { return t.CreatedAt }
func (t Target) GetUpdatedAt() time.Time  { return t.UpdatedAt }
func (t Target) GetLastActive() time.Time { return t.LastActive }
func (t Target) GetParentsUUID() map[RecordType]string {
	return map[RecordType]string{}
}

func (t Target) GetParentsTitle() map[RecordType]string {
	return map[RecordType]string{}
}

func (t Target) GetChildrenCount() int64 { return t.ActionsCount }

func (t Target) HasNote() bool { return t.HasNotes }

type Action struct {
	UUID          string       `json:"uuid"`
	CreatedAt     time.Time    `json:"created_at"`
	DueDate       sql.NullTime `json:"due_date"`
	UpdatedAt     time.Time    `json:"updated_at"`
	LastActive    time.Time    `json:"last_active"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	Notes         string       `json:"notes"`
	Status        string       `json:"status"`
	Version       int32        `json:"version"`
	TargetUUID    string       `json:"target_uuid"`
	TargetTitle   string       `json:"target_title"`
	HasNotes      bool         `json:"has_notes"`
	SessionsCount int64        `json:"sessions_count"`
}

func (a Action) ListItemView(hasSrc, chosen bool, width int) string {
	d := listItemData{
		title:  a.Title,
		status: a.Status,
	}
	if hasSrc {
		d.parent = a.TargetTitle
	}

	return listPageItemView(d, chosen, width)
}

func (a Action) ListItemDetailView(hasSrc bool, width int) string {
	var due *time.Time
	if a.DueDate.Valid {
		due = &a.DueDate.Time
	}

	d := listItemData{
		title:         a.Title,
		description:   a.Description,
		status:        a.Status,
		due:           due,
		itemType:      RecordTypeAction,
		hasNotes:      a.HasNotes,
		childrenCount: a.SessionsCount,
	}
	if hasSrc {
		d.parent = a.TargetTitle
	}

	return listPageItemDetail(d, width)
}

func (a Action) GetActualType() RecordType { return RecordTypeAction }
func (a Action) GetUUID() string           { return a.UUID }
func (a Action) GetTitle() string          { return a.Title }
func (a Action) GetDescription() string    { return a.Description }
func (a Action) GetStatus() string         { return a.Status }
func (a Action) GetNote() string           { return a.Notes }
func (a Action) GetDueDate() (time.Time, bool) {
	if a.DueDate.Valid {
		return a.DueDate.Time, true
	}
	return time.Time{}, false
}
func (a Action) GetCreatedAt() time.Time  { return a.CreatedAt }
func (a Action) GetUpdatedAt() time.Time  { return a.UpdatedAt }
func (a Action) GetLastActive() time.Time { return a.LastActive }
func (a Action) GetParentsUUID() map[RecordType]string {
	return map[RecordType]string{
		RecordTypeTarget: a.TargetUUID,
	}
}

func (a Action) GetParentsTitle() map[RecordType]string {
	return map[RecordType]string{
		RecordTypeTarget: a.TargetTitle,
	}
}

func (a Action) GetChildrenCount() int64 { return 0 }

func (a Action) HasNote() bool { return a.HasNotes }

type User struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type listItemData struct {
	title         string
	description   string
	status        string
	due           *time.Time
	parent        string
	itemType      RecordType
	hasNotes      bool
	childrenCount int64
}

func listPageItemView(d listItemData, chosen bool, width int) string {
	var stringBuilder strings.Builder

	if chosen {
		stringBuilder.WriteString(
			style.StatusTextStyle(d.status).MarginLeft(1).Render("∎") +
				style.ChoicesStyle["list"].Choice.Width(width-1).
					Margin(0, 1, 0, 0).
					Padding(0, 1, 0, 1).
					Render(d.title) + "\n",
		)
	} else {
		gap, remain, ok := style.CalculateGap(width-2, d.title, d.status)
		if !ok {
			gap = 1
		}

		stringBuilder.WriteString(
			style.StatusTextStyle(d.status).MarginLeft(1).Render("∎") +
				lipgloss.NewStyle().Width(width).Padding(0, 1).Render(
					style.ChoicesStyle["list"].Choices.Render(d.title)+
						strings.Repeat(" ", gap+remain)+
						style.StatusTextStyle(d.status).Render(strings.ToLower(d.status)), //+ "\n",
				) + "\n",
		)
	}

	return stringBuilder.String()
}

func listPageItemDetail(d listItemData, width int) string {
	var builder strings.Builder

	dueInfo := style.Document.Primary.Render("Due at: ")
	var dueValue string
	if d.due == nil {
		dueValue = "--"
	} else {
		dueValue = d.due.Format("2006-01-02")
	}
	dueValue = style.Document.Normal.Render(dueValue)

	statusInfo := style.Document.Primary.Render("Status: ")
	statusValue := style.StatusTextStyle(d.status).
		Render(strings.ToLower(d.status))

	notesInfo := style.Document.Primary.Render("Notes: ")
	var notesValue string
	if d.hasNotes {
		notesValue = lipgloss.NewStyle().Foreground(colors.Success).Render("✔")
	} else {
		notesValue = lipgloss.NewStyle().Foreground(colors.Danger).Render("✘")
	}

	var description string
	if d.description == "" {
		description = "---"
	} else {
		description = strings.TrimSpace(d.description)
	}

	builder.WriteString(
		style.Document.Primary.Padding(0, 2).Render("Description:") + "\n" +
			style.Document.Normal.Width(width).Padding(0, 2).Render(description) +
			"\n",
	)

	var childrenCountInfo string
	switch d.itemType {
	case RecordTypeTarget:
		childrenCountInfo = "Actions count: "
	case RecordTypeAction:
		childrenCountInfo = "Sessions count: "
	}

	if d.parent != "" {
		var parentInfo string
		switch d.itemType {
		case RecordTypeAction:
			parentInfo = style.Document.Primary.Render("Target: ")
		case RecordTypeSession:
			parentInfo = style.Document.Primary.Render("Action: ")
		}

		gap, remain, ok := style.CalculateGap(
			width-4,
			dueInfo+dueValue,
			statusInfo+statusValue,
			notesInfo+notesValue,
			fmt.Sprintf("%s: %d", childrenCountInfo, d.childrenCount),
		)
		if !ok {
			gap = 1
		}

		detail := lipgloss.NewStyle().Padding(0, 2).Render(
			parentInfo+style.Document.Normal.Render(d.parent),
		) + "\n"

		detail += lipgloss.NewStyle().Padding(0, 2).Render(
			dueInfo + dueValue +
				strings.Repeat(" ", gap+remain) +
				statusInfo + statusValue +
				strings.Repeat(" ", gap) +
				lipgloss.NewStyle().
					Render(
						style.Document.Primary.Render(childrenCountInfo)+
							style.Document.Normal.Render(
								fmt.Sprintf("%d", d.childrenCount),
							),
					) +
				strings.Repeat(" ", gap) +
				notesInfo + notesValue,
		)
		builder.WriteString(detail)
	} else {
		gap, remain, ok := style.CalculateGap(
			width-4,
			dueInfo+dueValue,
			statusInfo+statusValue,
			notesInfo+notesValue,
			fmt.Sprintf("%s: %d", childrenCountInfo, d.childrenCount),
		)
		if !ok {
			gap = 1
		}

		detail := lipgloss.NewStyle().Padding(0, 2).Render(
			dueInfo + dueValue +
				strings.Repeat(" ", gap+remain) +
				statusInfo + statusValue +
				strings.Repeat(" ", gap) +
				lipgloss.NewStyle().
					Render(
						style.Document.Primary.Render(childrenCountInfo)+
							style.Document.Normal.Render(
								fmt.Sprintf("%d", d.childrenCount),
							),
					) +
				strings.Repeat(" ", gap) +
				notesInfo + notesValue,
		)
		builder.WriteString(detail)
	}

	return builder.String()
}

func ListPageDetailView(detail string, dim bool) string {
	if dim {
		return lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colors.BorderMuted).
			Render(detail)
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colors.Text).
		Render(detail)
}
