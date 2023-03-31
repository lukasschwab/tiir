// Package render converts texts to varioius text formats: plaintext, HTML, and
// syndication feed representations of a collection of texts.
//
// Notably, this doens't include interactive interfaces for displaying new texts
// (e.g. at the command line).
package render

import (
	"io"

	"github.com/lukasschwab/tiir/pkg/text"
)

// Function rendering texts to the provided io.Writer.
type Function func(texts []*text.Text, to io.Writer) error

// TODO: standardize on either single-method interfaces *or* exported function
// types. https://eli.thegreenplace.net/2023/the-power-of-single-method-interfaces-in-go/
