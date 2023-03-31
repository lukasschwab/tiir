// Package text is the core data model: it's a text you read. Everything else is
// managing texts.
package text

import (
	"errors"
	"time"
)

// Text you read and recorded in this application.
type Text struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Author    string    `json:"author"`
	Note      string    `json:"note"`
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
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

// EditWith gets updates to t from the user with e.
func (t *Text) EditWith(e Editor) (final *Text, err error) {
	return e.Update(t)
}

// Editor for text. Update returns user updates to initial; it may modify
// initial in place, but that behavior isn't guaranteed.
type Editor interface {
	Update(initial *Text) (final *Text, err error)
}
