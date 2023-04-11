package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/lukasschwab/tiir/pkg/text"
)

// UseFile at path as a JSON store. If the file doesn't exist, it's created and
// initialized to an empty store.
//
// If you don't call [File.Close], the underlying [os.File] won't be closed.
func UseFile(path string) (Interface, error) {
	return useFile(path)
}

func useFile(path string) (*File, error) {
	db, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	f := &File{db: db}
	if err := f.load(); err != nil {
		return nil, fmt.Errorf("can't parse file contents: %w", err)
	}
	return f, nil
}

// File implements [Interface]; see [UseFile].
//
// Loads all texts from the underlying os.File into a memory store as an
// intermediate form for each operation; most of the store logic is delegated to
// the memory implementation.
type File struct {
	// Mutex for file handle operations.
	sync.Mutex
	db *os.File

	cache *Memory
}

// parse all records in f into memory.
func (f *File) load() error {
	f.Lock()
	defer f.Unlock()

	result := new(map[string]*text.Text)
	if _, err := f.db.Seek(0, 0); err != nil {
		return fmt.Errorf("couldn't seek to beginning of file before parsing: %v", err)
	} else if bytes, err := io.ReadAll(f.db); err != nil {
		return fmt.Errorf("couldn't read file: %w", err)
	} else if err := json.Unmarshal(bytes, result); err != nil {
		if len(bytes) == 0 {
			f.cache = useMemory()
			return nil
		}
		return fmt.Errorf("couldn't parse file JSON: %w", err)
	}

	f.cache = &Memory{texts: *result}
	return nil
}

// commit all records in memory to the underlying file. Overwrites everything.
func (f *File) commit() error {
	f.Lock()
	defer f.Unlock()

	if newContents, err := json.MarshalIndent(f.cache.texts, "", "\t"); err != nil {
		return fmt.Errorf("couldn't marshal texts to JSON: %w", err)
	} else if err := f.db.Truncate(0); err != nil {
		return fmt.Errorf("couldn't clear file before writing: %v", err)
	} else if _, err := f.db.Seek(0, 0); err != nil {
		return fmt.Errorf("couldn't seek to beginning of file after truncating: %v", err)
	} else if _, err = f.db.Write(newContents); err != nil {
		return fmt.Errorf("couldn't write to file: %w", err)
	}
	return nil
}

// Read implements [Interface].
func (f *File) Read(id string) (*text.Text, error) {
	return f.cache.Read(id)
}

// Upsert implements [Interface].
func (f *File) Upsert(t *text.Text) (*text.Text, error) {
	t, err := f.cache.Upsert(t)
	if err != nil {
		return nil, err
	} else if err := f.commit(); err != nil {
		return nil, err
	}
	return t, nil
}

// Delete implements [Interface].
func (f *File) Delete(id string) (*text.Text, error) {
	t, err := f.cache.Delete(id)
	if err != nil {
		return nil, err
	} else if err := f.commit(); err != nil {
		return nil, err
	}
	return t, nil
}

// List implements [Interface].
func (f *File) List(c text.Comparator, d text.Direction) ([]*text.Text, error) {
	return f.cache.List(c, d)
}

// Close implements [Interface].
func (f *File) Close() error {
	return f.db.Close()
}
