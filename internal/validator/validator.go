package validator

import (
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func MultipleValidators(validators ...func(string) error) func(string) error {
	return func(input string) error {
		for _, validate := range validators {
			if err := validate(input); err != nil {
				return err
			}
		}
		return nil
	}
}

var ValidFormats = []string{"2006-01-02"}

func ValidateDate(formats []string) func(string) error {
	return func(date string) error {
		if date == "" {
			return nil
		}

		for _, f := range formats {
			if _, err := time.ParseInLocation(f, date, time.Local); err == nil {
				return nil
			}
		}
		return errors.New("invalid format")
	}
}

func ValidateDateAfter(date time.Time) func(string) error {
	validFormats := []string{"2006-01-02"}

	return func(input string) error {
		if input == "" {
			return nil
		}

		var parsedDate time.Time
		var err error
		parsed := false
		for _, f := range validFormats {
			parsedDate, err = time.ParseInLocation(f, input, time.Local)
			if err == nil {
				parsed = true
				break
			}
		}
		if !parsed {
			return errors.New("invalid date format")
		}

		if parsedDate.Before(date) {
			return fmt.Errorf("must be after %s", date.Format("2006-01-02"))
		}
		return nil
	}
}

func ValidateRequired(msg string) func(string) error {
	return func(input string) error {
		if input == "" {
			return errors.New(msg)
		}
		return nil
	}
}

func ValidateReachMaxLength(length int) func(string) error {
	return func(input string) error {
		if utf8.RuneCountInString(input) == length {
			return fmt.Errorf("reached max length: %d", length)
		}
		return nil
	}
}

func ValidateMaxLength(length int) func(string) error {
	return func(input string) error {
		if utf8.RuneCountInString(input) > length {
			return fmt.Errorf("exceeds max length: %d", length)
		}
		return nil
	}
}

func ValidateMinLength(length int) func(string) error {
	return func(input string) error {
		if utf8.RuneCountInString(input) < length {
			return fmt.Errorf("too short")
		}
		return nil
	}
}

func ValidateEmail() func(string) error {
	emailRX := regexp.MustCompile(
		"^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$",
	)

	return func(input string) error {
		if !emailRX.MatchString(input) {
			return errors.New("invalid email format")
		}
		return nil
	}
}

func ValidatePasswordLength(min, max int) func(string) error {
	return func(input string) error {
		// Normalize password
		pw := norm.NFKC.String(input)

		if err := ValidateMinLength(min)(pw); err != nil {
			return err
		}
		if err := ValidateMaxLength(max)(pw); err != nil {
			return err
		}

		return nil
	}
}
