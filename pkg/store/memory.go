package store

import (
	"fmt"
	"sync"

	"github.com/lukasschwab/tiir/pkg/text"
)

func NewMemory(initialTexts ...*text.Text) Store {
	m := &memory{
		texts: make(map[string]*text.Text),
	}
	for _, t := range initialTexts {
		m.Create(t)
	}
	return m
}

// NOTE: a lot of our interaction is timestamp-based. Should this be an ordered
// structure instead of an ID-indexed one? Probably! Eliminates sort before
// templating.
type memory struct {
	sync.RWMutex
	texts map[string]*text.Text
}

func (m *memory) Create(t *text.Text) (*text.Text, error) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.texts[t.ID]; ok {
		return nil, fmt.Errorf("ID '%v' already exists", t.ID)
	}

	m.texts[t.ID] = t
	return t, nil
}

func (m *memory) Read(id string) (*text.Text, error) {
	m.RLock()
	defer m.RUnlock()

	text, ok := m.texts[id]
	if !ok {
		return nil, fmt.Errorf("no text with ID '%v'", id)
	}
	return text, nil
}

func (m *memory) Update(id string, new *text.Text) (*text.Text, error) {
	m.Lock()
	defer m.Unlock()

	updated, ok := m.texts[id]
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

	m.texts[id] = updated
	return updated, nil
}

func (m *memory) Delete(id string) (*text.Text, error) {
	m.Lock()
	defer m.Unlock()

	text, ok := m.texts[id]
	if !ok {
		return nil, fmt.Errorf("no text with ID '%v'", id)
	}

	delete(m.texts, id)
	return text, nil
}

func (m *memory) Close() error {
	return nil
}
