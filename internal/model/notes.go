package model

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
)

//go:embed templates/*.tmpl
var noteTemplates embed.FS

// func render(filename string, data any) ([]byte, error) {
func (m *NoteModel) render(data any) ([]byte, error) {
	// tpl, err := template.ParseFS(noteTemplates, "templates/"+filename)
	// if err != nil {
	// 	return nil, err
	// }
	tpl := template.Must(template.New("note").Parse(m.content))

	var b bytes.Buffer
	if err := tpl.Execute(&b, data); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

type (
	EditorFinishedMsg struct {
		err error
	}
)

type noteInfo struct {
	record  data.RecordType
	input   input
	parents map[data.RecordType]string
}

type NoteModel struct {
	note    *data.Note
	content string
	focused bool

	info      noteInfo
	infoCache noteInfo
	// record record
	// info struct {
	// 	record data.RecordType
	// 	// name string
	// 	input   input
	// 	parents map[data.RecordType]string
	// }

	err error
}

func NewNoteModel(
	recordType data.RecordType,
	r record,
	i input,
	// name string,
	// parents map[data.RecordType]string,
) *NoteModel {
	parents := map[data.RecordType]string{}
	if r != nil {
		parents = r.GetParentsTitle()
	}

	return &NoteModel{
		info: noteInfo{
			record:  recordType,
			input:   i,
			parents: parents,
		},
		// info: struct {
		// 	record data.RecordType
		// 	// name    string
		// 	input   input
		// 	parents map[data.RecordType]string
		// }{
		// 	record: recordType,
		// 	// name:    i.Value(),
		// 	input:   i,
		// 	parents: parents,
		// },
	}
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

		// m.updateInfoCache()

		// if len(content) == 0 {
		// 	if err := m.note.Remove(); err != nil {
		// 		m.err = errors.New("failed to remove empty note")
		// 		return m, nil
		// 	}
		// }
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

			content, err := m.note.Read()
			if err != nil {
				m.err = errors.New("failed to read note content")
				return m, nil
			}

			// Write default content if the note is empty
			if len(content) == 0 {
				c, err := m.defaultContent()
				if err != nil {
					m.err = errors.New(err.Error())
				}
				m.note.Write(c)
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
	data := struct {
		Name   string
		Target string
		Action string
	}{
		// Name:   m.info.name,
		Name:   m.info.input.Value(),
		Target: m.info.parents[data.RecordTypeTarget],
		Action: m.info.parents[data.RecordTypeAction],
	}
	content, _ := m.render(data)

	return string(content)
	// return m.content
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

func (m *NoteModel) defaultContent() ([]byte, error) {
	var filename string
	switch m.info.record {
	case data.RecordTypeTarget:
		filename = "note_target.tmpl"
	case data.RecordTypeAction:
		filename = "note_action.tmpl"
	case data.RecordTypeSession:
		filename = "note_session.tmpl"
	default:
		return []byte{}, nil
	}

	// data := struct {
	// 	Name   string
	// 	Target string
	// 	Action string
	// }{
	// 	// Name:   m.info.name,
	// 	Name:   m.info.input.Value(),
	// 	Target: m.info.parents[data.RecordTypeTarget],
	// 	Action: m.info.parents[data.RecordTypeAction],
	// }
	// content, err := render(filename, data)
	// if err != nil {
	// 	return []byte{}, err
	// }
	content, err := fs.ReadFile(noteTemplates, "templates/"+filename)
	if err != nil {
		return []byte{}, err
	}

	return content, nil
}

func (m *NoteModel) updateInfoCache() {
	m.infoCache.record = m.info.record
	m.infoCache.input = m.info.input
	m.infoCache.parents = make(map[data.RecordType]string)
	maps.Copy(m.infoCache.parents, m.info.parents)
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
