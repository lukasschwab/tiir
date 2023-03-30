package render

import (
	"fmt"
	"io"
	"time"

	"github.com/lukasschwab/go-jsonfeed"
	"github.com/lukasschwab/tiir/pkg/text"
)

type jsonFeed struct{}

func (j jsonFeed) Render(texts []*text.Text, to io.Writer) error {
	items := make([]jsonfeed.Item, len(texts))
	for i, text := range texts {
		item := jsonfeed.NewItem(text.ID)

		item.Title = text.Title
		item.URL = text.URL

		author := jsonfeed.NewAuthor()
		author.Name = item.Author
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
