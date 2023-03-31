// render texts to, well, text: plaintext, HTML, and syndication feed
// representations of a collection of texts. Notably, this doens't include
// interactive interfaces for displaying new texts (e.g. at the command line).
package render

import (
	"io"

	"github.com/lukasschwab/tiir/pkg/text"
)

// TODO: add an Atom feed, probably just using gorilla/feeds.
// TODO: Add an HTML feed using html/template.
// TODO: consider a tea-based renderer... but this doesn't adhere to the same
// interface; it necessarily renders to a terminal, not an io.Writer. Probably
// a special case in the CLI.
var (
	Plain    Renderer = plain{}
	JSONFeed          = jsonFeed{}
)

type Renderer interface {
	Render(texts []*text.Text, to io.Writer) error
}
