package edit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVim(t *testing.T) {
	assert.Implements(t, (*Editor)(nil), Vim{})
}
