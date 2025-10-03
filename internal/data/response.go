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

const (
	RecordTypeTarget  RecordType = "Target"
	RecordTypeAction  RecordType = "Action"
	RecordTypeSession RecordType = "Session"
)

type RecordType string

func (rt RecordType) GetParentType() RecordType {
	switch rt {
	case RecordTypeAction:
		return RecordTypeTarget
	case RecordTypeSession:
		return RecordTypeAction
	default:
		return ""
	}
}

type RecordParent struct {
	UUID  string
	Title string
}

type RecordParents map[RecordType]RecordParent

func (rp RecordParents) IsEmpty() bool {
	return len(rp) == 0
}

func (rp RecordParents) UUID(recordType RecordType) string {
	return rp[recordType].UUID
}

func (rp RecordParents) Title(recordType RecordType) string {
	return rp[recordType].Title
}

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
		d.parent = map[RecordType]string{RecordTypeTarget: a.TargetTitle}
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
		d.parent = map[RecordType]string{RecordTypeTarget: a.TargetTitle}
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

func (a Action) GetChildrenCount() int64 { return a.SessionsCount }

func (a Action) HasNote() bool { return a.HasNotes }

type Session struct {
	UUID        string       `json:"uuid"`
	StartsAt    time.Time    `json:"starts_at"`
	EndsAt      sql.NullTime `json:"ends_at"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Notes       string       `json:"notes"`
	Version     int32        `json:"version"`
	ActionUUID  string       `json:"action_uuid"`
	ActionTitle string       `json:"action_title"`
	TargetUUID  string       `json:"target_uuid"`
	TargetTitle string       `json:"target_title"`
	HasNotes    bool         `json:"has_notes"`
}

func (s Session) ListItemView(hasSrc, chosen bool, width int) string {
	d := listItemData{
		title:  s.GetTitle(),
		status: s.GetStatus(),
	}
	if hasSrc {
		d.parent = map[RecordType]string{
			RecordTypeTarget: s.TargetTitle,
			RecordTypeAction: s.ActionTitle,
		}
	}

	return listPageItemView(d, chosen, width)
}

func (s Session) ListItemDetailView(hasSrc bool, width int) string {
	d := listItemData{
		title:    s.GetTitle(),
		status:   s.GetStatus(),
		itemType: RecordTypeSession,
		hasNotes: s.HasNotes,
	}
	if hasSrc {
		d.parent = map[RecordType]string{
			RecordTypeTarget: s.TargetTitle,
			RecordTypeAction: s.ActionTitle,
		}
	}

	return listPageItemDetail(d, width)
}

func (s Session) GetActualType() RecordType { return RecordTypeSession }
func (s Session) GetUUID() string           { return s.UUID }
func (s Session) GetTitle() string {
	startStr := s.StartsAt.Local().Format("2006-01-02 15:04:05")
	endStr := "--"
	if s.EndsAt.Valid {
		endStr = s.EndsAt.Time.Local().Format("2006-01-02 15:04:05")
	}

	return fmt.Sprintf("%s → %s", startStr, endStr)
}
func (s Session) GetDescription() string { return "" }
func (s Session) GetStatus() string {
	if s.EndsAt.Valid {
		return "completed"
	}
	return "in progress"
}
func (s Session) GetNote() string               { return s.Notes }
func (s Session) GetDueDate() (time.Time, bool) { return time.Time{}, false }
func (s Session) GetCreatedAt() time.Time       { return s.CreatedAt }
func (s Session) GetUpdatedAt() time.Time       { return s.UpdatedAt }
func (s Session) GetLastActive() time.Time      { return s.UpdatedAt }
func (s Session) GetParentsUUID() map[RecordType]string {
	return map[RecordType]string{
		RecordTypeAction: s.ActionUUID,
		RecordTypeTarget: s.TargetUUID,
	}
}

func (s Session) GetParentsTitle() map[RecordType]string {
	return map[RecordType]string{
		RecordTypeAction: s.ActionTitle,
		RecordTypeTarget: s.TargetTitle,
	}
}
func (s Session) GetChildrenCount() int64 { return 0 }
func (s Session) HasNote() bool           { return s.HasNotes }

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
	parent        map[RecordType]string
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

	if d.itemType != RecordTypeSession {
		builder.WriteString(
			style.Document.Primary.Padding(0, 2).Render("Description:") + "\n" +
				style.Document.Normal.Width(width).Padding(0, 2).Render(description) +
				"\n",
		)
	}

	var childrenCountInfo string
	switch d.itemType {
	case RecordTypeTarget:
		childrenCountInfo = "Actions count: "
	case RecordTypeAction:
		childrenCountInfo = "Sessions count: "
	}

	var detail string
	if len(d.parent) != 0 {
		if d.itemType == RecordTypeAction || d.itemType == RecordTypeSession {
			detail += lipgloss.NewStyle().Padding(0, 2).Render(
				style.Document.Primary.Render("Target: ")+
					style.Document.Normal.Render(d.parent[RecordTypeTarget]),
			) + "\n"
		}
		if d.itemType == RecordTypeSession {
			detail += lipgloss.NewStyle().Padding(0, 2).Render(
				style.Document.Primary.Render("Action: ")+
					style.Document.Normal.Render(d.parent[RecordTypeAction]),
			) + "\n"
		}
	}

	fields := []string{
		dueInfo + dueValue,
		statusInfo + statusValue,
		notesInfo + notesValue,
	}
	if d.itemType != RecordTypeSession {
		fields = append(fields, fmt.Sprintf("%s: %d", childrenCountInfo, d.childrenCount))
	}

	gap, remain, ok := style.CalculateGap(width-4, fields...)
	if !ok {
		gap = 1
	}

	fieldsString := dueInfo + dueValue + strings.Repeat(" ", gap+remain) +
		statusInfo + statusValue
	if d.itemType != RecordTypeSession {
		fieldsString += strings.Repeat(" ", gap) +
			lipgloss.NewStyle().
				Render(
					style.Document.Primary.Render(childrenCountInfo)+
						style.Document.Normal.Render(
							fmt.Sprintf("%d", d.childrenCount),
						),
				)
	}
	fieldsString += strings.Repeat(" ", gap) + notesInfo + notesValue

	detail += lipgloss.NewStyle().Padding(0, 2).Render(
		fieldsString,
	)

	builder.WriteString(detail)

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
