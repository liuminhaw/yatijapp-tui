package model

import (
	"errors"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
)

type (
	EditorFinishedMsg struct {
		err error
	}
)

type NoteModel struct {
	note    *data.Note
	content string
	focused bool

	err error
}

func NewNoteModel() *NoteModel {
	return &NoteModel{}
}

func (m *NoteModel) Init() tea.Cmd {
	return nil
}

func (m *NoteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case EditorFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		content, err := m.note.Read()
		if err != nil {
			m.err = errors.New("failed to read note content")
			return m, nil
		}
		m.content = string(content)
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			if !m.focused {
				return m, nil
			}
			if m.note == nil {
				note, err := data.NewTempNote()
				if err != nil {
					m.err = errors.New("failed to create note")
					return m, nil
				}
				m.note = note
			}
			return m, openEditor(m.note.Path())
		}
	}

	return m, nil
}

func (m *NoteModel) View() string {
	// TODO: add view in glamour if needed in the future
	return ""
}

func (m *NoteModel) Focus() tea.Cmd {
	m.focused = true
	return nil
}

func (m *NoteModel) Focused() bool {
	return m.focused
}

func (m *NoteModel) Blur() {
	m.focused = false
}

func (m *NoteModel) Value() string {
	return m.content
}

func (m *NoteModel) SetValue(content string) error {
	if m.note == nil {
		note, err := data.NewTempNote()
		if err != nil {
			return err
		}
		m.note = note
	}

	if err := m.note.Write([]byte(content)); err != nil {
		return err
	}

	m.content = content
	return nil
}

func (m *NoteModel) Validate() {}
func (m *NoteModel) Error() string {
	if m.err != nil {
		return m.err.Error()
	}
	return ""
}

func openEditor(filepath string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if _, err := exec.LookPath("vim"); err == nil {
			editor = "vim"
		} else if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else {
			return func() tea.Msg {
				return EditorFinishedMsg{
					err: errors.New("no valid editor found, please set $EDITOR environment variable or install vim/nano"),
				}
			}
		}
	}

	c := exec.Command(editor, filepath)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		return EditorFinishedMsg{err}
	})
}
