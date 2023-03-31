package edit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTea(t *testing.T) {
	assert.Implements(t, (*Editor)(nil), Tea{})
}
