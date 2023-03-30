// text is the data model: it's a text you read. Everything else is managing
// texts.
package text

import (
	"errors"
	"time"
)

type Text struct {
	ID        string
	Title     string
	URL       string
	Author    string
	Note      string
	Timestamp time.Time
}

// Validate t is a sufficient text: user has provided all required fields.
func (t *Text) Validate() error {
	switch "" {
	case t.Title:
		return errors.New("must specify a title")
	case t.Author:
		return errors.New("must specify an author")
	case t.Note:
		return errors.New("must specify note")
	case t.URL:
		return errors.New("must specify ID")
	default:
		return nil
	}
}
