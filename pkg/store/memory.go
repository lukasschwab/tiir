package store

import (
	"fmt"
	"sync"

	"github.com/lukasschwab/tiir/pkg/text"
)

func UseMemory(initialTexts ...*text.Text) *memory {
	m := &memory{
		texts: make(map[string]*text.Text),
	}
	for _, t := range initialTexts {
		m.Upsert(t)
	}
	return m
}

type memory struct {
	sync.RWMutex
	texts map[string]*text.Text
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

func (m *memory) Upsert(t *text.Text) (*text.Text, error) {
	m.Lock()
	defer m.Unlock()

	m.texts[t.ID] = t
	return t, nil
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

func (m *memory) List(order text.Order) ([]*text.Text, error) {
	texts := make([]*text.Text, 0, len(m.texts))
	for _, t := range m.texts {
		texts = append(texts, t)
	}
	text.Sort(texts, order)
	return texts, nil
}

func (m *memory) Close() error {
	return nil
}
