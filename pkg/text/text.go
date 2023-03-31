// text is the data model: it's a text you read. Everything else is managing
// texts.
package text

import (
	"errors"
	"sort"
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

type comparator func(t1, t2 *Text) bool

type direction int

// You can sort texts by a comparator ascending or descending.
const (
	Ascending direction = iota
	Descending
)

type Order struct {
	Compare   comparator
	Direction direction
}

// Sort texts in-place in the specified order.
func Sort(texts []*Text, o Order) {
	sort.Sort(&textSorter{
		compare:   o.Compare,
		direction: o.Direction,
		texts:     texts,
	})
}

// textSorter implements sort.Interface for texts.
type textSorter struct {
	compare   comparator
	direction direction
	texts     []*Text
}

// Let implements sort.Interface.
func (ts *textSorter) Len() int {
	return len(ts.texts)
}

// Swap implements sort.Interface.
func (ts *textSorter) Swap(i, j int) {
	ts.texts[i], ts.texts[j] = ts.texts[j], ts.texts[i]
}

// Less implements sort.interface.
func (ts *textSorter) Less(i, j int) bool {
	if ts.direction == Ascending {
		return ts.compare(ts.texts[i], ts.texts[j])
	}
	return ts.compare(ts.texts[j], ts.texts[i])
}

// Timestamp is a
func Timestamps(t1, t2 *Text) bool {
	return t1.Timestamp.Before(t2.Timestamp)
}
