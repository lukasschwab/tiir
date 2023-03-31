package text

import "sort"

// Sort these texts.
type Sort []*Text

// By the provided comparator and direction, sort the texts in s in place.
func (s Sort) By(c Comparator, d Direction) {
	sort.Sort(&textSorter{Sort: s, compare: c, direction: d})
}

// A Comparator determines which of two texts is "lesser" for sorting.
type Comparator func(t1, t2 *Text) bool

// A Direction determines whether to apply a comparator in [Ascending] or
// [Descending] order.
type Direction int

// You can sort texts in Ascending or Descending direction.
const (
	Ascending Direction = iota
	Descending
)

// textSorter implements sort.Interface for texts.
type textSorter struct {
	Sort
	compare   Comparator
	direction Direction
}

// Let implements sort.Interface.
func (ts *textSorter) Len() int {
	return len(ts.Sort)
}

// Swap implements sort.Interface.
func (ts *textSorter) Swap(i, j int) {
	ts.Sort[i], ts.Sort[j] = ts.Sort[j], ts.Sort[i]
}

// Less implements sort.interface.
func (ts *textSorter) Less(i, j int) bool {
	if ts.direction == Ascending {
		return ts.compare(ts.Sort[i], ts.Sort[j])
	}
	// For descending sort, invert the comparator.
	return ts.compare(ts.Sort[j], ts.Sort[i])
}

// Timestamps is a Comparator: sort texts by the time they were created.
func Timestamps(t1, t2 *Text) bool {
	return t1.Timestamp.Before(t2.Timestamp)
}
