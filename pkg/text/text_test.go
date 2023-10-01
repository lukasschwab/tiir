package text

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	assert.Error(t, (&Text{}).Validate())

	assert.Error(t, (&Text{Author: "a", Note: "n"}).Validate())
	assert.Error(t, (&Text{Author: "a", URL: "u"}).Validate())
	assert.Error(t, (&Text{Note: "n", URL: "u"}).Validate())

	assert.NoError(t, (&Text{Author: "a", Note: "n", URL: "u", Title: "t"}).Validate())
}

func TestRandomID(t *testing.T) {
	set := map[string]bool{}
	for i := 0; i < 10; i++ {
		id, err := RandomID()
		assert.NoError(t, err)
		assert.Len(t, id, 8)
		assert.NotContains(t, set, id)
		set[id] = true
	}
}
