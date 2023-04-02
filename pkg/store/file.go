package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/lukasschwab/tiir/pkg/text"
)

// UseFile uses the file at path as a JSON store. If the file doesn't exist,
// it's created and initialized to an empty store.
//
// If you don't call (*File).Close, the underlying [os.File] won't be closed.
func UseFile(path string) (Store, error) {
	return useFile(path)
}

func useFile(path string) (*file, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	fStore := &file{File: f}
	if i, err := fStore.parse(); err != nil {
		return nil, fmt.Errorf("can't parse file contents: %w", err)
	} else if err := fStore.commit(i); err != nil {
		return nil, fmt.Errorf("can't write to file: %w", err)
	}
	return fStore, nil
}

// file implementation of Store.
//
// Loads all texts from the underlying os.File into a memory store as an
// intermediate form for each operation; most of the store logic is delegated to
// the memory implementation.
//
// NOTE: simultaneous writes, e.g. from simultaneous HTTP requests, can yield
// unexpected behavior: every write overwrites the full file. This could be
// improved by persisting a centralized `memory` here to consolidate writes.
type file struct {
	sync.Mutex
	*os.File
}

// parse all records in f into memory.
func (f *file) parse() (*memory, error) {
	f.Lock()
	defer f.Unlock()

	result := new(map[string]*text.Text)
	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("couldn't seek to beginning of file before parsing: %v", err)
	} else if bytes, err := io.ReadAll(f.File); err != nil {
		return nil, fmt.Errorf("couldn't read file: %w", err)
	} else if err := json.Unmarshal(bytes, result); err != nil {
		if len(bytes) == 0 {
			return useMemory(), nil
		}
		return nil, fmt.Errorf("couldn't parse file JSON: %w", err)
	}
	return &memory{texts: *result}, nil
}

// commit all records in memory to the underlying file. Overwrites everything.
func (f *file) commit(new *memory) error {
	f.Lock()
	defer f.Unlock()

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

// Read implements Store.
func (f *file) Read(id string) (*text.Text, error) {
	m, err := f.parse()
	if err != nil {
		return nil, err
	}

	return m.Read(id)
}

// Upsert implements Store.
func (f *file) Upsert(t *text.Text) (*text.Text, error) {
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

// Delete implements Store.
func (f *file) Delete(id string) (*text.Text, error) {
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

// List implements Store.
func (f *file) List(c text.Comparator, d text.Direction) ([]*text.Text, error) {
	m, err := f.parse()
	if err != nil {
		return nil, err
	}

	return m.List(c, d)
}

// Close implements Store.
func (f *file) Close() error {
	return f.File.Close()
}
