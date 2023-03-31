package store

import (
	"fmt"
	"sync"

	"github.com/lukasschwab/tiir/pkg/text"
)

// UseMemory constructs a new in-memory store containing initialTexts.
//
// Because it doesn't persist texts after the program terminates, memory stores
// are best suited for testing and for intermediate formats used by other
// Stores.
func UseMemory(initialTexts ...*text.Text) Store {
	return useMemory(initialTexts...)
}

func useMemory(initialTexts ...*text.Text) *memory {
	m := &memory{
		texts: make(map[string]*text.Text),
	}
	for _, t := range initialTexts {
		m.Upsert(t)
	}
	return m
}

// memory implementation of Store.
type memory struct {
	sync.RWMutex
	texts map[string]*text.Text
}

// Read implements Store.
func (m *memory) Read(id string) (*text.Text, error) {
	m.RLock()
	defer m.RUnlock()

	text, ok := m.texts[id]
	if !ok {
		return nil, fmt.Errorf("no text with ID '%v'", id)
	}
	return text, nil
}

// Upsert implements Store.
func (m *memory) Upsert(t *text.Text) (*text.Text, error) {
	m.Lock()
	defer m.Unlock()

	m.texts[t.ID] = t
	return t, nil
}

// Delete implements Store.
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

// List implements Store.
func (m *memory) List(c text.Comparator, d text.Direction) ([]*text.Text, error) {
	texts := make([]*text.Text, 0, len(m.texts))
	for _, t := range m.texts {
		texts = append(texts, t)
	}

	text.Sort(texts).By(c, d)
	return texts, nil
}

// Close implements Store.
func (m *memory) Close() error {
	return nil
}
