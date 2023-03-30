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
var (
	Plain    Renderer = plain{}
	JSONFeed          = jsonFeed{}
)

// TODO: interface should support ordering, limit parameters.
type Renderer interface {
	Render(texts []*text.Text, to io.Writer) error
}
