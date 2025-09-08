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

func (t Target) ListItemView(chosen bool, width int) string {
	var stringBuilder strings.Builder

	if chosen {
		dueInfo := style.ChoicesStyle["list"].Choices.Bold(true).Render("Due at:")
		var dueValue string
		if !t.DueDate.Valid {
			dueValue = style.ChoicesStyle["list"].Choices.Render("--")
		} else {
			dueValue = style.ChoicesStyle["list"].Choices.Render(t.DueDate.Time.Format("2006-01-02"))
		}
		statusInfo := style.ChoicesStyle["list"].Choices.Bold(true).Render("Status:")
		statusValue := style.ChoicesStyle["list"].Choices.Render(strings.ToUpper(t.Status))
		hr := lipgloss.NewStyle().Foreground(colors.HelperText).
			Render("  " + strings.Repeat("─", width-2))

		stringBuilder.WriteString(
			style.ChoicesStyle["list"].Choice.
				Width(width).
				Render(fmt.Sprintf(" ‣ %s", t.Title)) + "\n",
		)
		stringBuilder.WriteString(hr + "\n")

		var description string
		if t.Description == "" {
			description = "---"
		} else {
			description = strings.TrimSpace(t.Description)
		}

		stringBuilder.WriteString(
			style.ChoicesStyle["list"].ChoiceContent.
				Width(width).
				Render(fmt.Sprintf("  %s", description)) + "\n",
		)
		stringBuilder.WriteString(
			fmt.Sprintf(
				"  %s %s  |  %s %s\n",
				dueInfo,
				dueValue,
				statusInfo,
				statusValue,
			),
		)
		stringBuilder.WriteString(hr + "\n")
	} else {
		stringBuilder.WriteString(
			style.ChoicesStyle["default"].Choices.Width(width).Render(fmt.Sprintf("- %s", t.Title)) + "\n",
		)
	}

	return stringBuilder.String()
}

type User struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
