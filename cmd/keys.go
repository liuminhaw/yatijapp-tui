package main

import (
	"errors"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func isFnKey(msg tea.KeyMsg) bool {
	s := msg.String()

	if !strings.HasPrefix(s, "f") {
		return false
	}
	n, err := strconv.Atoi(s[1:])
	if err != nil || n < 1 || n > 12 {
		return false
	}

	return true
}

func fnKeyNumber(msg tea.KeyMsg) (int, error) {
	if !isFnKey(msg) {
		return 0, errors.New("not a function key")
	}

	s := msg.String()
	n, err := strconv.Atoi(s[1:])
	if err != nil || n < 1 || n > 12 {
		return 0, errors.New("not a function key")
	}

	return n, nil
}
