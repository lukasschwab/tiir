package text

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	assert.Error(t, (&Text{}).Validate())

	assert.Error(t, (&Text{Author: "a", Note: "n"}).Validate())
	assert.Error(t, (&Text{Author: "a", URL: "u"}).Validate())
	assert.Error(t, (&Text{Note: "n", URL: "u"}).Validate())

	assert.NoError(t, (&Text{Author: "a", Note: "n", URL: "u", Title: "t"}).Validate())
}

func TestSort(t *testing.T) {
	texts := []*Text{
		{Title: "earliest", Timestamp: time.Date(-1, 0, 0, 0, 0, 0, 0, time.UTC)},
		{Title: "middle", Timestamp: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)},
		{Title: "latest", Timestamp: time.Date(1, 0, 0, 0, 0, 0, 0, time.UTC)},
	}

	Sort(texts, Order{Compare: Timestamps, Direction: Ascending})
	assert.Equal(t, "earliest", texts[0].Title)
	assert.Equal(t, "middle", texts[1].Title)
	assert.Equal(t, "latest", texts[2].Title)

	Sort(texts, Order{Compare: Timestamps, Direction: Descending})
	assert.Equal(t, "latest", texts[0].Title)
	assert.Equal(t, "middle", texts[1].Title)
	assert.Equal(t, "earliest", texts[2].Title)
}
