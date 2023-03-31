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
