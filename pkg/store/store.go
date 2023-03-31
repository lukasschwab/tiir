package store

import (
	"github.com/lukasschwab/tiir/pkg/text"
)

// TODO: implement CRUD.
// Key question: does Store own the tir.text structure?
type Store interface {
	// Read a text by ID.
	Read(id string) (*text.Text, error)
	// Delete a text by ID and return the deleted text.
	Delete(id string) (*text.Text, error)
	// Upsert a text by t.ID and return the resulting text. Assumes t.ID is set
	// and t is valid; see (*text.Text).Validate(...).
	Upsert(t *text.Text) (*text.Text, error)
	// TODO: introduce an ordered List function for supporting templates.

	List(order text.Order) ([]*text.Text, error)

	// Close implements io.Closer.
	Close() error
}
