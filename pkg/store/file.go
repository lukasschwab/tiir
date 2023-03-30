package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/lukasschwab/tiir/pkg/text"
)

// TODO: provide a constructor?

// File driven implementation of Store.
type File struct {
	*os.File
}

type inner map[string]*text.Text

func (f *File) parse() (inner, error) {
	result := new(inner)
	if bytes, err := io.ReadAll(f.File); err != nil {
		return nil, fmt.Errorf("couldn't read file: %w", err)
	} else if err := json.Unmarshal(bytes, result); err != nil {
		return nil, fmt.Errorf("couldn't parse file JSON: %w", err)
	}
	return *result, nil
}

func (f *File) commit(new inner) error {
	if newContents, err := json.MarshalIndent(new, "", "\t"); err != nil {
		return fmt.Errorf("couldn't marshal texts to JSON: %w", err)
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

	updated, ok := texts[new.ID]
	if !ok {
		return nil, fmt.Errorf("no text with ID '%v'", new.ID)
	}

	if new.Author != "" {
		updated.Author = new.Author
	}
	if new.Note != "" {
		updated.Note = new.Note
	}
	if new.URL != "" {
		updated.URL = new.URL
	}

	texts[new.ID] = updated

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
