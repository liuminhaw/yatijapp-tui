package data

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type ErrorResponse struct {
	Error any `json:"error,omitempty"`
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
