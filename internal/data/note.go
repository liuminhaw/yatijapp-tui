package data

import (
	"os"
	"path"
)

type Note struct {
	// dir  string
	file *os.File
}

func NewTempNote(subdirs ...string) (*Note, error) {
	tmpDir := path.Join(os.TempDir(), "yatijapp-tui", "notes", path.Join(subdirs...))
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, err
	}

	file, err := os.CreateTemp(tmpDir, "*.md")
	if err != nil {
		return nil, err
	}

	return &Note{
		// dir:  tmpDir,
		file: file,
	}, nil
}

func (n *Note) Path() string {
	return n.file.Name()
}

func (n *Note) Read() ([]byte, error) {
	data, err := os.ReadFile(n.file.Name())
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (n *Note) Write(content []byte) error {
	if _, err := n.file.Write(content); err != nil {
		return err
	}
	return nil
}

func (n *Note) ReadOnly() error {
	return n.file.Chmod(0440)
}
