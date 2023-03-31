// render texts to, well, text: plaintext, HTML, and syndication feed
// representations of a collection of texts. Notably, this doens't include
// interactive interfaces for displaying new texts (e.g. at the command line).
package render

import (
	"io"

	"github.com/lukasschwab/tiir/pkg/text"
)

type Function func(texts []*text.Text, to io.Writer) error
