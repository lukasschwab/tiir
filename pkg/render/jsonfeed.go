package render

import (
	"fmt"
	"io"
	"time"

	"github.com/lukasschwab/go-jsonfeed"
	"github.com/lukasschwab/tiir/pkg/text"
)

// JSONFeed rendering for texts. Example output:
//
//	{
//		"version": "https://jsonfeed.org/version/1",
//		"title": "tir",
//		"items": [
//			{
//				"id": "35bb8126",
//				"url": "https://davidchall.github.io/ggip/articles/visualizing-ip-data.html",
//				"title": "Visualizing IP data",
//				"content_text": "Use a Hilbert Curve: efficient 2D packing that keeps consecutive sequences spatially contiguous.",
//				"date_published": "2023-04-07T21:43:52-07:00",
//				"authors": [
//	 				{
//						"name": "David Hall"
//					}
//				]
//			}
//		]
//	}
func JSONFeed(texts []*text.Text, to io.Writer) error {
	items := make([]jsonfeed.Item, len(texts))
	for i, text := range texts {
		item := jsonfeed.NewItem(text.ID)

		item.Title = text.Title
		item.URL = text.URL

		author := jsonfeed.NewAuthor()
		author.Name = text.Author
		item.Authors = []jsonfeed.Author{author}

		item.ContentText = text.Note
		item.DatePublished = text.Timestamp.Format(time.RFC3339)

		items[i] = item
	}

	if bytes, err := jsonfeed.NewFeed("tir", items).ToJSON(); err != nil {
		return fmt.Errorf("error converting feed to JSON: %w", err)
	} else if _, err := to.Write(bytes); err != nil {
		return fmt.Errorf("error writing feed: %w", err)
	}
	return nil
}
