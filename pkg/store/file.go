package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/lukasschwab/tiir/pkg/text"
)

func UseFile(path string) (*File, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	f := &File{File: file}
	if i, err := f.parse(); err != nil {
		return nil, fmt.Errorf("can't parse file contents: %w", err)
	} else if err := f.commit(i); err != nil {
		return nil, fmt.Errorf("can't write to file: %w", err)
	}
	return f, nil
}

// File driven implementation of Store.
type File struct {
	*os.File
}

func (f *File) parse() (*memory, error) {
	result := new(map[string]*text.Text)
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("couldn't seek to beginning of file before parsing: %v", err)
	} else if bytes, err := io.ReadAll(f.File); err != nil {
		return nil, fmt.Errorf("couldn't read file: %w", err)
	} else if err := json.Unmarshal(bytes, result); err != nil {
		if len(bytes) == 0 {
			return UseMemory(), nil
		}
		return nil, fmt.Errorf("couldn't parse file JSON: %w", err)
	}
	return &memory{texts: *result}, nil
}

func (f *File) commit(new *memory) error {
	if newContents, err := json.MarshalIndent(new.texts, "", "\t"); err != nil {
		return fmt.Errorf("couldn't marshal texts to JSON: %w", err)
	} else if err := f.Truncate(0); err != nil {
		return fmt.Errorf("couldn't clear file before writing: %v", err)
	} else if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("couldn't seek to beginning of file after truncating: %v", err)
	} else if _, err = f.Write(newContents); err != nil {
		return fmt.Errorf("couldn't write to file: %w", err)
	}
	return nil
}

func (f *File) Read(id string) (*text.Text, error) {
	m, err := f.parse()
	if err != nil {
		return nil, err
	}

	return m.Read(id)
}

func (f *File) Upsert(t *text.Text) (*text.Text, error) {
	m, err := f.parse()
	if err != nil {
		return nil, err
	}

	if t, err = m.Upsert(t); err != nil {
		return nil, err
	} else if err := f.commit(m); err != nil {
		return nil, err
	}

	return t, nil
}

func (f *File) Delete(id string) (*text.Text, error) {
	m, err := f.parse()
	if err != nil {
		return nil, err
	}

	t, err := m.Delete(id)
	if err != nil {
		return nil, err
	} else if err := f.commit(m); err != nil {
		return nil, err
	}
	return t, nil
}

func (f *File) List(order text.Order) ([]*text.Text, error) {
	m, err := f.parse()
	if err != nil {
		return nil, err
	}

	return m.List(order)
}

func (f *File) Close() error {
	return f.File.Close()
}
