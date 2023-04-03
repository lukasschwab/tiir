// Package tir provides the interface-facing CRUD service.
package tir

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
)

const (
	// idLength for hexadecimal text IDs.
	idLength = 8
)

// New constructs a new Service around s.
func New(s store.Interface) *Service {
	return &Service{provider: s}
}

// Service for managing tir texts.
//
// TODO: don't expose Service's internals; expose an interface.
type Service struct {
	provider store.Interface
}

// Create a text.
func (s *Service) Create(text *text.Text) (*text.Text, error) {
	var err error
	if err = text.Validate(); err != nil {
		return nil, err
	} else if text.ID, err = randomID(); err != nil {
		return nil, fmt.Errorf("couldn't randomize ID: %w", err)
	}
	text.Timestamp = time.Now()

	return s.provider.Upsert(text)
}

// Read a text by ID.
func (s *Service) Read(id string) (*text.Text, error) {
	return s.provider.Read(id)
}

// Update a text by ID and return the resulting text.
func (s *Service) Update(id string, updates *text.Text) (*text.Text, error) {
	extant, err := s.provider.Read(id)
	if err != nil {
		return nil, fmt.Errorf("error reading old record: %w", err)
	}
	// Don't validate: updates can be partial.
	extant.Integrate(updates)
	return s.provider.Upsert(extant)
}

// Delete a text by ID and return the deleted text.
func (s *Service) Delete(id string) (*text.Text, error) {
	return s.provider.Delete(id)
}

// List all texts available to the service.
func (s *Service) List() ([]*text.Text, error) {
	return s.provider.List(text.Timestamps, text.Descending)
}

// Close the underlying Store.
func (s *Service) Close() error {
	return s.provider.Close()
}

func randomID() (string, error) {
	bytes := make([]byte, idLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
