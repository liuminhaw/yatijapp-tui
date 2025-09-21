package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

const (
	// Format in output: ✓ option\t - option\t - option\t
	RadioDefaultView = iota
	RadioListView
)

type Radio struct {
	options  []string
	choice   string
	selected int
	focused  bool
	viewType int
	width    int
}

func NewRadio(options []string, viewType, width int) Radio {
	return Radio{
		options:  options,
		choice:   "",
		selected: 0,
		focused:  false,
		viewType: viewType,
		width:    width,
	}
}

func (r *Radio) Blur() {
	r.focused = false
}

func (r *Radio) Focus() tea.Cmd {
	r.focused = true
	return nil
}

func (r Radio) Focused() bool {
	return r.focused
}

func (r Radio) Update(msg tea.Msg) (Radio, tea.Cmd) {
	if !r.focused {
		return r, nil
	}

	switch r.viewType {
	case RadioListView:
		return r.listViewUpdate(msg)
	default:
		return r.defaultUpdate(msg)
	}
}

func (r Radio) View() string {
	switch r.viewType {
	case RadioListView:
		return r.listView()
	default:
		return r.defaultView()
	}
}

func (r Radio) Selected() string {
	return r.options[r.selected]
}

func (r Radio) Value() string {
	if r.choice != "" {
		return r.choice
	}
	return r.options[r.selected]
}

func (r *Radio) SetValue(value string) error {
	for i, option := range r.options {
		if option == value {
			r.selected = i
			r.choice = value
			return nil
		}
	}
	return fmt.Errorf("value %s not found in options", value)
}

// func (r Radio) Options() []string {
// 	return r.options
// }

func (r Radio) defaultUpdate(msg tea.Msg) (Radio, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			r.selected--
			if r.selected < 0 {
				r.selected = len(r.options) - 1
			}
		case "right", "l":
			r.selected++
			if r.selected >= len(r.options) {
				r.selected = 0
			}
		case "enter":
			r.choice = r.options[r.selected]
			return r, nil // Handle selection logic here if needed
		}
	}
	return r, nil
}

func (r Radio) listViewUpdate(msg tea.Msg) (Radio, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			for {
				r.selected--
				if r.selected < 0 {
					r.selected = len(r.options) - 1
				}
				if r.options[r.selected] != "" {
					break
				}
			}
		case "down", "j":
			for {
				r.selected++
				if r.selected >= len(r.options) {
					r.selected = 0
				}
				if r.options[r.selected] != "" {
					break
				}
			}
		case "enter":
			r.choice = r.options[r.selected]
			return r, nil // Handle selection logic here if needed
		}
	}
	return r, nil
}

func (r Radio) defaultView() string {
	options := make([]string, len(r.options))
	for i, option := range r.options {
		if i == r.selected {
			options[i] = style.ChoicesStyle["default"].Choice.Render(
				"✓ " + option,
			)
		} else {
			options[i] = style.ChoicesStyle["default"].Choices.Render(
				"- " + option,
			)
		}
	}

	gap, remain, ok := style.CalculateGap(r.width, options...)
	if !ok {
		gap = 1
		remain = 0
	}

	var builder strings.Builder
	for i, option := range options {
		builder.WriteString(option)
		if i < len(options)-1 {
			builder.WriteString(strings.Repeat(" ", gap))
			if remain > 0 {
				builder.WriteString(" ")
				remain--
			}
		}
	}

	return builder.String()
}

func (r Radio) listView() string {
	options := make([]string, len(r.options))
	for i, option := range r.options {
		if i == r.selected {
			options[i] = style.ChoicesStyle["list"].Choice.Width(26).Padding(0, 1).Render(
				fmt.Sprintf("%s", option),
			)
		} else {
			options[i] = style.ChoicesStyle["list"].Choices.Width(26).Padding(0, 1).Render(
				fmt.Sprintf("%s", option),
			)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, options...)
}
