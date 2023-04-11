// Package store stores a [text.Text] collection somewhere. Typically, use the
// [github.com/lukasschwab/tiir/pkg/tir.Interface] provided by
// [github.com/lukasschwab/tiir/pkg/config.Load], rather than constructing a
// store directly, to use the user-configured store.
//
// An initialized store must be closed: call Close when you're done
// writing to the [Interface].
package store

import (
	"io"

	"github.com/lukasschwab/tiir/pkg/text"
)

// Interface for storing texts somewhere. An initialized store must be closed:
// call Close when you're done writing to the store.
type Interface interface {
	// Close the Store, rendering it unusable for future operations.
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
