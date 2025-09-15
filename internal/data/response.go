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
	case RecordTypeActivity:
		return RecordTypeTarget, true
	case RecordTypeSession:
		return RecordTypeActivity, true
	default:
		return "", false
	}
}

const (
	RecordTypeTarget   RecordType = "Target"
	RecordTypeActivity RecordType = "Activity"
	RecordTypeSession  RecordType = "Session"
)

type Target struct {
	UUID        string       `json:"uuid"`
	CreatedAt   time.Time    `json:"created_at"`
	DueDate     sql.NullTime `json:"due_date"`
	UpdatedAt   time.Time    `json:"updated_at"`
	LastActive  time.Time    `json:"last_active"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Notes       string       `json:"notes"`
	Status      string       `json:"status"`
	Version     int32        `json:"version"`
}

func (t Target) ListItemView(hasSrc, chosen bool, width int) string {
	var due *time.Time
	if t.DueDate.Valid {
		due = &t.DueDate.Time
	}
	return listPageItemView(listItemData{
		title:       t.Title,
		description: t.Description,
		status:      t.Status,
		due:         due,
	}, chosen, width)
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

type Activity struct {
	UUID        string       `json:"uuid"`
	CreatedAt   time.Time    `json:"created_at"`
	DueDate     sql.NullTime `json:"due_date"`
	UpdatedAt   time.Time    `json:"updated_at"`
	LastActive  time.Time    `json:"last_active"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Notes       string       `json:"notes"`
	Status      string       `json:"status"`
	Version     int32        `json:"version"`
	TargetUUID  string       `json:"target_uuid"`
	TargetTitle string       `json:"target_title"`
}

func (a Activity) ListItemView(hasSrc, chosen bool, width int) string {
	var due *time.Time
	if a.DueDate.Valid {
		due = &a.DueDate.Time
	}

	d := listItemData{
		title:       a.Title,
		description: a.Description,
		status:      a.Status,
		due:         due,
	}
	if hasSrc {
		d.target = a.TargetTitle
	}

	return listPageItemView(d, chosen, width)
}

func (a Activity) GetActualType() RecordType { return RecordTypeActivity }
func (a Activity) GetUUID() string           { return a.UUID }
func (a Activity) GetTitle() string          { return a.Title }
func (a Activity) GetDescription() string    { return a.Description }
func (a Activity) GetStatus() string         { return a.Status }
func (a Activity) GetNote() string           { return a.Notes }
func (a Activity) GetDueDate() (time.Time, bool) {
	if a.DueDate.Valid {
		return a.DueDate.Time, true
	}
	return time.Time{}, false
}
func (a Activity) GetCreatedAt() time.Time  { return a.CreatedAt }
func (a Activity) GetUpdatedAt() time.Time  { return a.UpdatedAt }
func (a Activity) GetLastActive() time.Time { return a.LastActive }
func (a Activity) GetParentsUUID() map[RecordType]string {
	return map[RecordType]string{
		RecordTypeTarget: a.TargetUUID,
	}
}

func (a Activity) GetParentsTitle() map[RecordType]string {
	return map[RecordType]string{
		RecordTypeTarget: a.TargetTitle,
	}
}

type User struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type listItemData struct {
	title       string
	description string
	status      string
	due         *time.Time
	target      string
}

func listPageItemView(d listItemData, chosen bool, width int) string {
	var stringBuilder strings.Builder

	if chosen {
		separator := style.ChoicesStyle["list"].Choices.Bold(true).Render("|")
		dueInfo := style.ChoicesStyle["list"].Choices.Bold(true).Render("Due at:")
		var dueValue string
		if d.due == nil {
			dueValue = style.ChoicesStyle["list"].Choices.Render("--")
		} else {
			dueValue = style.ChoicesStyle["list"].Choices.Render(d.due.Format("2006-01-02"))
		}
		statusInfo := style.ChoicesStyle["list"].Choices.Bold(true).Render("Status:")
		statusValue := style.StatusTextStyle(d.status).Render(strings.ToLower(d.status))
		hr := lipgloss.NewStyle().Foreground(colors.HelperText).
			Render("  " + strings.Repeat("─", width-2))

		stringBuilder.WriteString(
			lipgloss.NewStyle().Bold(true).Foreground(colors.DocumentText).Render("▸ ") +
				style.ChoicesStyle["list"].Choice.
					Width(width-2).
					Render(" "+d.title) + "\n",
		)
		stringBuilder.WriteString(hr + "\n")

		var description string
		if d.description == "" {
			description = "---"
		} else {
			description = strings.TrimSpace(d.description)
		}

		stringBuilder.WriteString(
			style.ChoicesStyle["list"].ChoiceContent.
				Width(width).
				Render("  "+description) + "\n",
		)
		stringBuilder.WriteString(
			"  " + dueInfo + " " + dueValue + "  " + separator + "   " + statusInfo + " " + statusValue + "\n",
		)

		if d.target != "" {
			targetInfo := style.ChoicesStyle["list"].Choices.Bold(true).Render("Target:")
			targetValue := style.ChoicesStyle["list"].Choices.Render(d.target)
			stringBuilder.WriteString(
				"  " + targetInfo + " " + targetValue + "\n",
			)
		}
		stringBuilder.WriteString(hr + "\n")
	} else {
		stringBuilder.WriteString(
			style.ChoicesStyle["default"].Choices.Width(width).Render("▫ "+d.title) + "\n",
		)
	}

	return stringBuilder.String()
}
