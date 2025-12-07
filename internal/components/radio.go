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

func (c *Radio) Blur() {
	c.focused = false
}

func (c *Radio) Focus() tea.Cmd {
	c.focused = true
	return nil
}

func (c Radio) Focused() bool {
	return c.focused
}

func (c Radio) Update(msg tea.Msg) (Radio, tea.Cmd) {
	if !c.focused {
		return c, nil
	}

	switch c.viewType {
	case RadioListView:
		return c.listViewUpdate(msg)
	default:
		return c.defaultUpdate(msg)
	}
}

func (c Radio) View() string {
	switch c.viewType {
	case RadioListView:
		return c.listView()
	default:
		return c.defaultView()
	}
}

func (c Radio) Selected() string {
	return c.options[c.selected]
}

func (r Radio) Value() string {
	if r.choice != "" {
		return r.choice
	}
	return r.options[r.selected]
}

func (c *Radio) SetValue(value string) error {
	for i, option := range c.options {
		if option == value {
			c.selected = i
			c.choice = value
			return nil
		}
	}
	return fmt.Errorf("value %s not found in options", value)
}

func (c Radio) defaultUpdate(msg tea.Msg) (Radio, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			c.selected--
			if c.selected < 0 {
				c.selected = len(c.options) - 1
			}
		case "right", "l":
			c.selected++
			if c.selected >= len(c.options) {
				c.selected = 0
			}
		case "enter":
			c.choice = c.options[c.selected]
			return c, nil // Handle selection logic here if needed
		}
	}
	return c, nil
}

func (c Radio) listViewUpdate(msg tea.Msg) (Radio, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			for {
				c.selected--
				if c.selected < 0 {
					c.selected = len(c.options) - 1
				}
				if c.options[c.selected] != "" {
					break
				}
			}
		case "down", "j":
			for {
				c.selected++
				if c.selected >= len(c.options) {
					c.selected = 0
				}
				if c.options[c.selected] != "" {
					break
				}
			}
		case "enter":
			c.choice = c.options[c.selected]
			return c, nil // Handle selection logic here if needed
		}
	}
	return c, nil
}

func (c Radio) defaultView() string {
	options := make([]string, len(c.options))
	for i, option := range c.options {
		options[i] = choiceFormat(i == c.selected, i == c.selected, c.focused, option)
	}

	gap, remain, ok := style.CalculateGap(c.width, options...)
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

func (c Radio) listView() string {
	options := make([]string, len(c.options))
	for i, option := range c.options {
		if i == c.selected {
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
