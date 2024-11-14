// Package text is the core data model: it's a text you read (past tense).
// All the other packages in this module modify or manage texts.
package text

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

const (
	IDLength = 2 * 2 * 2
)

// Text you read and recorded in this application.
type Text struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Author    string    `json:"author"`
	Note      string    `json:"note"`
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Public    bool      `json:"public"`
}

// Validate t has nonzero values for all required fields:
//
//   - Title
//   - Author
//   - Note
//   - URL
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
// initial in place, but that behavior isn't guaranteed. See implementations in
// [pkg/edit].
type Editor interface {
	Update(initial *Text) (final *Text, err error)
}

// Integrate updates into t in-place, skipping zero-value update fields (empty-
// string authors, for example) and non-user-editable fields (e.g.
// Timestamp).
func (t *Text) Integrate(updates *Text) {
	if updates.Author != "" {
		t.Author = updates.Author
	}
	if updates.Note != "" {
		t.Note = updates.Note
	}
	if updates.Title != "" {
		t.Title = updates.Title
	}
	if updates.URL != "" {
		t.URL = updates.URL
	}
}

func RandomID() (string, error) {
	bytes := make([]byte, IDLength/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
