package render

import (
	"encoding/json"
	"io"

	"github.com/lukasschwab/tiir/pkg/text"
)

// JSON list renderer for texts. Example output:
//
//	[{
//		"title": "Visualizing IP data",
//		"url": "https://davidchall.github.io/ggip/articles/visualizing-ip-data.html",
//		"author": "David Hall",
//		"note": "Use a Hilbert Curve: efficient 2D packing that keeps consecutive sequences spatially contiguous.",
//		"id": "35bb8126",
//		"timestamp": "2023-04-07T21:43:52.776451-07:00"
//	}]
func JSON(texts []*text.Text, to io.Writer) error {
	return json.NewEncoder(to).Encode(texts)
}
