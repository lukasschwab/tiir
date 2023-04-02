package render

import (
	"encoding/json"
	"io"

	"github.com/lukasschwab/tiir/pkg/text"
)

// JSON list renderer for texts.
func JSON(texts []*text.Text, to io.Writer) error {
	return json.NewEncoder(to).Encode(texts)
}
