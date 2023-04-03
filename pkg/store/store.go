package store

import (
	"io"

	"github.com/lukasschwab/tiir/pkg/text"
)

// Interface for storing texts somewhere. An initialized store must be closed:
// call or defer (Interface).Close.
type Interface interface {
	// Close closes the Store, rendering it unusable for future operations.
	io.Closer
	// Read a text by ID.
	Read(id string) (*text.Text, error)
	// Delete a text by ID and return the deleted text.
	Delete(id string) (*text.Text, error)
	// Upsert a text by t.ID and return the resulting text. Assumes t.ID is set
	// and t is valid; see (*text.Text).Validate(...).
	Upsert(t *text.Text) (*text.Text, error)
	// List all texts in the store in order.
	List(c text.Comparator, d text.Direction) ([]*text.Text, error)
}
