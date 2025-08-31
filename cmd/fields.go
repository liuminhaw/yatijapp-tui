package main

import (
	"errors"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/liuminhaw/yatijapp-tui/internal/model"
	"github.com/liuminhaw/yatijapp-tui/internal/validator"
)

func emailField(width int, focus bool) Focusable {
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

func passwordField(width int, focus bool) Focusable {
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

func usernameField(width int, focus bool) Focusable {
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

func tokenField(width int, focus bool, placeholder string) Focusable {
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

func fieldsValidation(fields []Focusable, msg string) error {
	for _, field := range fields {
		field.Validate()
		if err := field.Error(); err != "" {
			return errors.New(msg)
		}
	}

	return nil
}
