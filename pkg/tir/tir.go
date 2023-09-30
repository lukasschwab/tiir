// Package tir provides the interface-facing CRUD service.
package tir

import (
	"fmt"
	"io"
	"time"

	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
)

// Interface for managing tir texts. Callers should use this in lieu of
// store.Store; the latter assumes application-level validation implemented in
// this package. See [New].
type Interface interface {
	io.Closer
	// Create a new text. This function is responsible for assigning
	// [text.Text.ID] and [text.Text.Timestamp].
	Create(new *text.Text) (*text.Text, error)
	// Read a text by its ID.
	Read(id string) (*text.Text, error)
	// Update a text with ID to include updates. Zero-valued fields in updates
	// (e.g. empty-string fields) are ignored.
	Update(id string, updates *text.Text) (*text.Text, error)
	// Delete a text by its ID.
	Delete(id string) (*text.Text, error)
	// List all texts sorted by decreasing [text.Text.Timestamp].
	List() ([]*text.Text, error)
}

// New constructs a new application [Interface] around s. In general, use
// [github.com/lukasschwab/tiir/pkg/config.Load] instead to respect user
// configuration.
func New(s store.Interface) Interface {
	return &app{provider: s}
}

// app for managing tir texts.
type app struct {
	provider store.Interface
}

// Create a text.
func (s *app) Create(t *text.Text) (*text.Text, error) {
	var err error
	if err = t.Validate(); err != nil {
		return nil, err
	} else if t.ID, err = text.RandomID(); err != nil {
		return nil, fmt.Errorf("couldn't randomize ID: %w", err)
	}
	if t.Timestamp.IsZero() {
		t.Timestamp = time.Now()
	}
	return s.provider.Upsert(t)
}

// Read a text by ID.
func (s *app) Read(id string) (*text.Text, error) {
	return s.provider.Read(id)
}

// Update a text by ID and return the resulting text.
func (s *app) Update(id string, updates *text.Text) (*text.Text, error) {
	extant, err := s.provider.Read(id)
	if err != nil {
		return nil, fmt.Errorf("error reading old record: %w", err)
	}
	// Don't validate: updates can be partial.
	extant.Integrate(updates)
	return s.provider.Upsert(extant)
}

// Delete a text by ID and return the deleted text.
func (s *app) Delete(id string) (*text.Text, error) {
	return s.provider.Delete(id)
}

// List all texts available to the service.
func (s *app) List() ([]*text.Text, error) {
	return s.provider.List(text.Timestamps, text.Descending)
}

// Close the underlying Store.
func (s *app) Close() error {
	return s.provider.Close()
}
