package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSort(t *testing.T) {
	texts := []*Text{
		{Title: "earliest", Timestamp: time.Date(-1, 0, 0, 0, 0, 0, 0, time.UTC)},
		{Title: "middle", Timestamp: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)},
		{Title: "latest", Timestamp: time.Date(1, 0, 0, 0, 0, 0, 0, time.UTC)},
	}

	Sort(texts).By(Timestamps, Ascending)
	assert.Equal(t, "earliest", texts[0].Title)
	assert.Equal(t, "middle", texts[1].Title)
	assert.Equal(t, "latest", texts[2].Title)

	Sort(texts).By(Timestamps, Descending)
	assert.Equal(t, "latest", texts[0].Title)
	assert.Equal(t, "middle", texts[1].Title)
	assert.Equal(t, "earliest", texts[2].Title)
}
