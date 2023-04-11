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
func UseMemory(initialTexts ...*text.Text) Interface {
	return useMemory(initialTexts...)
}

func useMemory(initialTexts ...*text.Text) *Memory {
	m := &Memory{
		texts: make(map[string]*text.Text),
	}
	for _, t := range initialTexts {
		m.Upsert(t)
	}
	return m
}

// Memory implements [Interface] in-memory. See [UseMemory].
type Memory struct {
	sync.RWMutex
	texts map[string]*text.Text
}

// Read implements [Interface].
func (m *Memory) Read(id string) (*text.Text, error) {
	m.RLock()
	defer m.RUnlock()

	text, ok := m.texts[id]
	if !ok {
		return nil, fmt.Errorf("no text with ID '%v'", id)
	}
	return text, nil
}

// Upsert implements [Interface].
func (m *Memory) Upsert(t *text.Text) (*text.Text, error) {
	m.Lock()
	defer m.Unlock()

	m.texts[t.ID] = t
	return t, nil
}

// Delete implements [Interface].
func (m *Memory) Delete(id string) (*text.Text, error) {
	m.Lock()
	defer m.Unlock()

	text, ok := m.texts[id]
	if !ok {
		return nil, fmt.Errorf("no text with ID '%v'", id)
	}

	delete(m.texts, id)
	return text, nil
}

// List implements [Interface].
func (m *Memory) List(c text.Comparator, d text.Direction) ([]*text.Text, error) {
	texts := make([]*text.Text, 0, len(m.texts))
	for _, t := range m.texts {
		texts = append(texts, t)
	}

	text.Sort(texts).By(c, d)
	return texts, nil
}

// Close implements [Interface].
func (m *Memory) Close() error {
	return nil
}
