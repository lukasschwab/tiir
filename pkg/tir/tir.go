// package tir provides the interface-facing CRUD service.
package tir

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
)

// TODO: receive something implementing store.Store.
func New(s store.Store) *Service {
	return &Service{Store: s}
}

type Service struct {
	Store store.Store
}

func (s *Service) Create(text *text.Text) (*text.Text, error) {
	if err := text.Validate(); err != nil {
		return nil, err
	}

	text.ID = toID(text)
	text.Timestamp = time.Now()

	return s.Store.Upsert(text)
}

func (s *Service) Read(id string) (*text.Text, error) {
	return s.Store.Read(id)
}

func (s *Service) Update(id string, updates *text.Text) (*text.Text, error) {
	applied, err := s.Store.Read(id)
	if err != nil {
		return nil, fmt.Errorf("error reading old record: %w", err)
	}

	// Apply changes to the extant record.
	if updates.Author != "" {
		applied.Author = updates.Author
	}
	if updates.Note != "" {
		applied.Note = updates.Note
	}
	if updates.Title != "" {
		applied.Title = updates.Title
	}
	if updates.URL != "" {
		applied.URL = updates.URL
	}

	// Don't validate: updates can be partial.
	return s.Store.Upsert(applied)
}

func (s *Service) Delete(id string) (*text.Text, error) {
	return s.Store.Delete(id)
}

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
