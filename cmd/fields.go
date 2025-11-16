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

func nameInput(width int, focus bool, source string) Focusable {
	maxLen := 80
	field := textinput.New()
	field.Prompt = ""
	field.Placeholder = "Give " + source + " a name"
	field.Width = width - 1
	field.CharLimit = maxLen
	field.Validate = validator.MultipleValidators(
		validator.ValidateRequired("required"),
		validator.ValidateMaxLength(maxLen),
	)
	if focus {
		field.Focus()
	}

	return model.NewTextInputWrapper(field)
}

func dueInput(width int, focus bool, validators ...func(string) error) Focusable {
	field := textinput.New()
	field.Prompt = ""
	field.Placeholder = "YYYY-mm-dd"
	field.Width = width - 1

	v := []func(string) error{validator.ValidateDateTime(validator.ValidDateFormats)}
	for _, validator := range validators {
		v = append(v, validator)
	}
	field.Validate = validator.MultipleValidators(v...)
	if focus {
		field.Focus()
	}

	return model.NewTextInputWrapper(field)
}

func descriptionField(width int, focus bool) Focusable {
	lenMax := 200
	field := textinput.New()
	field.Prompt = ""
	field.Placeholder = "Information about the record"
	field.Width = width - 1
	field.CharLimit = lenMax
	field.Validate = validator.ValidateReachMaxLength(lenMax)
	if focus {
		field.Focus()
	}

	return model.NewTextInputWrapper(field)
}

func timeInput(width int, focus bool) Focusable {
	field := textinput.New()
	field.Prompt = ""
	field.Placeholder = "YYYY-mm-dd HH:MM:SS"
	field.Width = width - 1
	field.Validate = validator.ValidateDateTime(validator.ValidDateTimeFormats)
	if focus {
		field.Focus()
	}

	return model.NewTextInputWrapper(field)
}

func emailInput(width int, focus bool) Focusable {
	field := textinput.New()
	field.Prompt = ""
	field.Placeholder = "email"
	field.Width = width - 1
	field.Validate = validator.MultipleValidators(
		validator.ValidateRequired("required"),
		validator.ValidateEmail(),
	)
	if focus {
		field.Focus()
	}

	return model.NewTextInputWrapper(field)
}

func passwordInput(width int, focus bool) Focusable {
	field := textinput.New()
	field.EchoMode = textinput.EchoPassword
	field.Prompt = ""
	field.Placeholder = "password"
	field.Width = width - 1
	field.Validate = validator.MultipleValidators(
		validator.ValidateRequired("required"),
		validator.ValidatePasswordLength(8, 72),
	)
	if focus {
		field.Focus()
	}

	return model.NewTextInputWrapper(field)
}

func passwordConfirmField(width int, focus bool, match *model.TextInputWrapper) Focusable {
	field := textinput.New()
	field.EchoMode = textinput.EchoPassword
	field.Prompt = ""
	field.Placeholder = "confirm password"
	field.Width = width - 1
	field.Validate = validator.MultipleValidators(
		validator.ValidateRequired("required"),
		model.ValidateTextInputMatch(match, "password not match"),
	)
	if focus {
		field.Focus()
	}

	return model.NewTextInputWrapper(field)
}

func usernameInput(width int, focus bool) Focusable {
	field := textinput.New()
	field.Focus()
	field.Prompt = ""
	field.Placeholder = "username"
	field.Width = width - 1
	field.Validate = validator.MultipleValidators(
		validator.ValidateRequired("required"),
		validator.ValidateMinLength(2),
		validator.ValidateMaxLength(30),
	)
	if focus {
		field.Focus()
	}

	return model.NewTextInputWrapper(field)
}

func tokenInput(width int, focus bool, placeholder string) Focusable {
	field := textinput.New()
	field.Focus()
	field.Prompt = ""
	field.Placeholder = placeholder
	field.Width = width - 1
	field.Validate = validator.ValidateRequired("required")
	if focus {
		field.Focus()
	}

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
