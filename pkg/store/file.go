package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/lukasschwab/tiir/pkg/text"
)

// TODO: provide a constructor?
func OpenOrCreateFile(path string) (*File, error) {
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

type inner map[string]*text.Text

func (f *File) parse() (inner, error) {
	result := new(inner)
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("couldn't seek to beginning of file before parsing: %v", err)
	} else if bytes, err := io.ReadAll(f.File); err != nil {
		return nil, fmt.Errorf("couldn't read file: %w", err)
	} else if err := json.Unmarshal(bytes, result); err != nil {
		if len(bytes) == 0 {
			return map[string]*text.Text{}, nil
		}
		return nil, fmt.Errorf("couldn't parse file JSON: %w", err)
	}
	return *result, nil
}

func (f *File) commit(new inner) error {
	if newContents, err := json.MarshalIndent(new, "", "\t"); err != nil {
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

func (f *File) Create(t *text.Text) (*text.Text, error) {
	texts, err := f.parse()
	if err != nil {
		return nil, err
	}

	texts[t.ID] = t

	if err := f.commit(texts); err != nil {
		return nil, err
	}
	return t, nil
}

func (f *File) Read(id string) (*text.Text, error) {
	texts, err := f.parse()
	if err != nil {
		return nil, err
	}

	text, ok := texts[id]
	if !ok {
		return nil, fmt.Errorf("no text with ID '%v'", id)
	}
	return text, nil
}

func (f *File) Update(id string, new *text.Text) (*text.Text, error) {
	texts, err := f.parse()
	if err != nil {
		return nil, err
	}

	texts[id] = new

	if err := f.commit(texts); err != nil {
		return nil, err
	}
	return new, nil
}

func (f *File) Delete(id string) (*text.Text, error) {
	texts, err := f.parse()
	if err != nil {
		return nil, err
	}

	text, ok := texts[id]
	if !ok {
		return nil, fmt.Errorf("no text with ID '%v'", id)
	}

	delete(texts, id)
	if err := f.commit(texts); err != nil {
		return nil, err
	}
	return text, nil
}

func (f *File) Close() error {
	return f.File.Close()
}
