package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

func choiceFormat(selected, checked, focused bool, content string) string {
	var bullet string
	if selected && focused {
		bullet = "➨ "
	} else {
		bullet = "- "
	}

	var bulletStyle lipgloss.Style
	if focused {
		bulletStyle = style.InputStyle.Selected
	} else if checked && selected {
		bulletStyle = style.ChoicesStyle["default"].Choice
	} else {
		bulletStyle = style.ChoicesStyle["default"].Choices
	}

	var formatted string
	if checked {
		if selected {
			formatted = bulletStyle.Bold(true).Render(bullet) +
				style.ChoicesStyle["default"].Choice.Render(content)
		} else {
			formatted = style.ChoicesStyle["default"].Choice.Render(bullet + content)
		}
	} else {
		if selected {
			formatted = bulletStyle.Render(bullet) +
				style.ChoicesStyle["default"].Choices.Render(content)
		} else {
			formatted = style.ChoicesStyle["default"].Choices.Render(bullet + content)
		}
	}

	return formatted
}
