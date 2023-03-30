package tir

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"

	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
)

// TODO: receive something implementing store.Store.
func New(s store.Store) *Service {
	return &Service{store: s}
}

type Service struct {
	store store.Store
}

func (s *Service) Create(text *text.Text) (*text.Text, error) {
	if err := text.Validate(); err != nil {
		return nil, err
	}

	text.ID = toID(text)
	text.Timestamp = time.Now()

	return s.store.Create(text)
}

func (s *Service) Read(id string) (*text.Text, error) {
	return s.store.Read(id)
}

func (s *Service) Update(id string, new *text.Text) (*text.Text, error) {
	// Prevent confusion.
	new.ID = id
	// Don't validate: updates can be partial.
	return s.store.Update(id, new)
}

func (s *Service) Delete(id string) (*text.Text, error) {
	return s.store.Delete(id)
}

// NOTE: should this really be random?
func toID(text *text.Text) string {
	h := md5.New()
	for _, elem := range []string{text.URL, text.Author, text.Note} {
		io.WriteString(h, elem)
	}
	return fmt.Sprintf("%x", h.Sum(nil)[:8])
}
