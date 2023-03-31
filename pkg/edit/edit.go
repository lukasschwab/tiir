// Package edit provides interfaces for editing texts. Typically these are
// command line interfaces: a cusotm CLI [Tea] or JSON in the user's preferred
// text editor.
package edit

import "github.com/lukasschwab/tiir/pkg/text"

// Editor for text. Update returns user updates to initial; it may modify
// initial in place, but that behavior isn't guaranteed.
type Editor interface {
	Update(initial *text.Text) (final *text.Text, err error)
}
