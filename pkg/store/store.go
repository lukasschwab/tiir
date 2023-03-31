package store

import (
	"github.com/lukasschwab/tiir/pkg/text"
)

// TODO: implement CRUD.
// Key question: does Store own the tir.text structure?
type Store interface {
	// Create a new text. This function assumes t.ID and t.Timestamp are already
	// set by the caller.
	Create(t *text.Text) (*text.Text, error)
	// Read a text by ID.
	Read(id string) (*text.Text, error)
	// Update a text by ID. Empty strings in the new text are skipped.
	Update(id string, new *text.Text) (*text.Text, error)
	// Delete a text by ID and return the deleted text.
	Delete(id string) (*text.Text, error)
	// TODO: introduce an ordered List function for supporting templates.

	// Close implements io.Closer.
	Close() error
}
