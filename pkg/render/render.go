// Package render converts texts to varioius text formats: plaintext, HTML, and
// syndication feed representations of a collection of texts. Notably, this
// doesn't include interactive interfaces for displaying new texts (e.g. at the
// command line).
package render

import (
	"io"

	"github.com/lukasschwab/tiir/pkg/text"
)

// Function rendering texts to the provided io.Writer. Use this to extend this
// app with new renderers, or use one of the provided implementations:
//
//   - [Plain]
//   - [JSON]
//   - [JSONFeed]
//   - [HTML]
type Function func(texts []*text.Text, to io.Writer) error
