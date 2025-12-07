package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
)

type checkboxOption struct {
	label   string
	checked bool
}

type Checkbox struct {
	options  []checkboxOption
	focused  bool
	selected int
	width    int
}

func NewCheckbox(options []string, width int) Checkbox {
	checkboxOptions := make([]checkboxOption, len(options))
	for i, option := range options {
		checkboxOptions[i] = checkboxOption{
			label:   option,
			checked: false,
		}
	}

	return Checkbox{
		options:  checkboxOptions,
		focused:  false,
		selected: 0,
		width:    width,
	}
}

func (c *Checkbox) Blur() {
	c.focused = false
}

func (c *Checkbox) Focus() tea.Cmd {
	c.focused = true
	return nil
}

func (c Checkbox) Focused() bool {
	return c.focused
}

func (c Checkbox) Update(msg tea.Msg) (Checkbox, tea.Cmd) {
	if !c.focused {
		return c, nil
	}

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
		case " ":
			c.options[c.selected].checked = !c.options[c.selected].checked
		}
	}
	return c, nil
}

func (c Checkbox) View() string {
	options := make([]string, len(c.options))
	for i, option := range c.options {
		options[i] = choiceFormat(i == c.selected, option.checked, c.focused, option.label)
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

func (c Checkbox) Values() []string {
	selected := []string{}
	for _, option := range c.options {
		if option.checked {
			selected = append(selected, option.label)
		}
	}

	return selected
}

// func (c Checkbox) SetValue(values []string) error {
func (c Checkbox) SetValues(values ...string) error {
outer:
	for _, value := range values {
		for i, option := range c.options {
			if option.label == value {
				c.options[i].checked = true
				continue outer
			}
		}
		return fmt.Errorf("value %s not found in options", value)
	}

	return nil
}
