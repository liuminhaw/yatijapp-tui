package main

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/liuminhaw/yatijapp-tui/colors"
	"github.com/liuminhaw/yatijapp-tui/internal/model"
	"github.com/liuminhaw/yatijapp-tui/internal/style"
	"github.com/liuminhaw/yatijapp-tui/internal/validator"
)

type inputFieldConfig struct {
	width       int
	focus       bool
	placeholder string
	lenMax      int // optional max length limit, 0 means no limit
	isSecret    bool
	validators  []func(string) error
}

func nameInput(width int, focus bool, recordType string) Focusable {
	c := inputFieldConfig{
		width:       width,
		focus:       focus,
		placeholder: "Give " + recordType + " a name",
		lenMax:      80,
		validators: []func(string) error{
			validator.ValidateRequired("required"),
			validator.ValidateMaxLength(80),
		},
	}

	return generalInput(c)
}

func dueInput(width int, focus bool, validators ...func(string) error) Focusable {
	c := inputFieldConfig{
		width:       width,
		focus:       focus,
		placeholder: "YYYY-mm-dd",
		lenMax:      0,
		validators: []func(string) error{
			validator.ValidateDateTime(validator.ValidDateFormats),
		},
	}

	for _, validator := range validators {
		c.validators = append(c.validators, validator)
	}

	return generalInput(c)
}

func descriptionInput(width int, focus bool) Focusable {
	c := inputFieldConfig{
		width:       width,
		focus:       focus,
		placeholder: "Information about the record",
		lenMax:      200,
		validators: []func(string) error{
			validator.ValidateReachMaxLength(200),
		},
	}

	return generalInput(c)
}

func timeInput(width int, focus bool) Focusable {
	c := inputFieldConfig{
		width:       width,
		focus:       focus,
		placeholder: "YYYY-mm-dd HH:MM:SS",
		lenMax:      0,
		validators: []func(string) error{
			validator.ValidateDateTime(validator.ValidDateTimeFormats),
		},
	}

	return generalInput(c)
}

func emailInput(width int, focus bool) Focusable {
	c := inputFieldConfig{
		width:       width,
		focus:       focus,
		placeholder: "email",
		lenMax:      0,
		validators: []func(string) error{
			validator.ValidateRequired("required"),
			validator.ValidateEmail(),
		},
	}

	return generalInput(c)
}

func passwordInput(width int, focus bool) Focusable {
	c := inputFieldConfig{
		width:       width,
		focus:       focus,
		placeholder: "password",
		lenMax:      0,
		isSecret:    true,
		validators: []func(string) error{
			validator.ValidateRequired("required"),
			validator.ValidatePasswordLength(8, 72),
		},
	}

	return generalInput(c)
}

func passwordConfirmInput(width int, focus bool, match *model.TextInputWrapper) Focusable {
	c := inputFieldConfig{
		width:       width,
		focus:       focus,
		placeholder: "confirm password",
		lenMax:      0,
		isSecret:    true,
		validators: []func(string) error{
			validator.ValidateRequired("required"),
			model.ValidateTextInputMatch(match, "password not match"),
		},
	}

	return generalInput(c)
}

func usernameInput(width int, focus bool) Focusable {
	c := inputFieldConfig{
		width:       width,
		focus:       focus,
		placeholder: "username",
		lenMax:      30,
		validators: []func(string) error{
			validator.ValidateRequired("required"),
			validator.ValidateMinLength(2),
			validator.ValidateMaxLength(30),
		},
	}

	return generalInput(c)
}

func generalInput(c inputFieldConfig) Focusable {
	field := textinput.New()
	field.Prompt = ""
	field.Placeholder = c.placeholder
	field.Width = c.width - 1

	if c.isSecret {
		field.EchoMode = textinput.EchoPassword
	}
	if c.focus {
		field.Focus()
	}
	if c.lenMax > 0 {
		field.CharLimit = c.lenMax
	}
	field.Validate = validator.MultipleValidators(c.validators...)

	return model.NewTextInputWrapper(field)
}

func inputsValidation(inputs []Focusable, msg string) error {
	for _, input := range inputs {
		input.Validate()
		if err := input.Error(); err != "" {
			return errors.New(msg)
		}
	}

	return nil
}

type field struct {
	idx int
	obj Focusable
}

func (f field) prompt(focusedPrompt, blurredPrompt string) string {
	var prompt string
	if f.obj.Focused() {
		prompt = focusedPrompt
	} else {
		prompt = blurredPrompt
	}

	return prompt
}

func (f field) simpleTitlePrompt(
	name, helper string,
	withErr bool,
	styling ...lipgloss.Style,
) string {
	var prompt string
	if !withErr {
		prompt = fmt.Sprintf(
			"%s %s",
			style.FormFieldStyle.Prompt(name, f.obj.Focused()),
			style.FormFieldStyle.Helper.Render(helper),
		)
	} else {
		prompt = fmt.Sprintf(
			"%s %s\n%s",
			style.FormFieldStyle.Prompt(name, f.obj.Focused()),
			style.FormFieldStyle.Helper.Render(helper),
			style.FormFieldStyle.Error.Render(f.obj.Error()),
		)
	}

	appliedStyle := lipgloss.NewStyle()
	for _, s := range styling {
		appliedStyle = s.Inherit(appliedStyle)
	}

	return appliedStyle.Render(prompt)
}

func (f field) textInputPrompt(name, helper string, styling ...lipgloss.Style) string {
	prompt := fmt.Sprintf(
		"%s %s\n%s\n%s",
		style.FormFieldStyle.Prompt(name, f.obj.Focused()),
		style.FormFieldStyle.Helper.Render(helper),
		style.FormFieldStyle.Content.Render(f.obj.View()),
		style.FormFieldStyle.Error.Render(f.obj.Error()),
	)

	appliedStyle := lipgloss.NewStyle()
	for _, s := range styling {
		appliedStyle = s.Inherit(appliedStyle)
	}

	return appliedStyle.Render(prompt)
}

func (f field) selectionPrompt(
	parent, helper string,
	compact bool,
	styling ...lipgloss.Style,
) string {
	prompt := style.FormFieldStyle.Prompt(parent+" Source", f.obj.Focused()) +
		" " + style.FormFieldStyle.Helper.Render(helper)

	var input string
	if f.obj.View() == "" {
		input = lipgloss.NewStyle().Foreground(colors.HelperTextDim).Render(parent + " name")
	} else {
		input = style.FormFieldStyle.Content.Render(f.obj.View())
	}

	applyStyling := lipgloss.NewStyle().Width(formWidth).Margin(1, 5, 0)
	for _, s := range styling {
		applyStyling = s.Inherit(applyStyling)
	}

	if compact {
		return applyStyling.Render(
			prompt + " " + style.FormFieldStyle.Error.Render(f.obj.Error()) + "\n" + input,
		)
	} else {
		return applyStyling.Render(
			prompt + "\n" + input + "\n" + style.FormFieldStyle.Error.Render(f.obj.Error()),
		)
	}
}
