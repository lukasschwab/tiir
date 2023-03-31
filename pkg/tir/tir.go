// Package tir provides the interface-facing CRUD service.
package tir

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
)

// FromConfig loads a tir.Service from defaults, overridden by user-provided
// configuration.
//
// TODO: actually use a config!
func FromConfig() (*Service, error) {
	if home, err := os.UserHomeDir(); err != nil {
		return nil, fmt.Errorf("error getting user home directory: %v", err)
	} else if store, err := store.UseFile(home + "/.tir.json"); err != nil {
		return nil, fmt.Errorf("error opening tir file: %v", err)
	} else {
		return &Service{Store: store}, nil
	}
}

// New constructs a new Service around s.
func New(s store.Store) *Service {
	return &Service{Store: s}
}

// Service for managing tir texts.
//
// TODO: don't expose Service's internals; expose an interface.
type Service struct {
	Store store.Store
}

// Create a text.
func (s *Service) Create(text *text.Text) (*text.Text, error) {
	if err := text.Validate(); err != nil {
		return nil, err
	}

	text.ID = toID(text)
	text.Timestamp = time.Now()

	return s.Store.Upsert(text)
}

// Read a text by ID.
func (s *Service) Read(id string) (*text.Text, error) {
	return s.Store.Read(id)
}

// Update a text by ID and return teh resulting text.
func (s *Service) Update(id string, updates *text.Text) (*text.Text, error) {
	extant, err := s.Store.Read(id)
	if err != nil {
		return nil, fmt.Errorf("error reading old record: %w", err)
	}
	// Don't validate: updates can be partial.
	extant.Integrate(updates)
	return s.Store.Upsert(extant)
}

// Delete a text by ID and return the deleted text.
func (s *Service) Delete(id string) (*text.Text, error) {
	return s.Store.Delete(id)
}

// List all texts available to the service.
func (s *Service) List() ([]*text.Text, error) {
	// TODO; parameterize the sort order.
	return s.Store.List(text.Timestamps, text.Descending)
}

// Close the underlying Store.
func (s *Service) Close() error {
	return s.Store.Close()
}

// NOTE: should this really be random?
func toID(text *text.Text) string {
	h := md5.New()
	for _, elem := range []string{text.Title, text.URL, text.Author, text.Note} {
		if _, err := io.WriteString(h, elem); err != nil {
			log.Fatalf("Couldn't hash text element: %v", err)
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:8]
}
