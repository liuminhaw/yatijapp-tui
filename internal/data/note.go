package data

import (
	"os"
	"path"
)

type Note struct {
	// dir  string
	file *os.File
}

func NewTempNote() (*Note, error) {
	tmpDir := path.Join(os.TempDir(), "yatijapp-tui", "notes")
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
	_, err := n.file.Seek(0, 0)
	if err != nil {
		return err
	}

	if _, err := n.file.Write(content); err != nil {
		return err
	}
	return nil
}

func (n *Note) Remove() error {
	if err := os.Remove(n.file.Name()); err != nil {
		return err
	}
	return nil
}
