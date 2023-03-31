package edit

import (
	"testing"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"
)

func TestVim(t *testing.T) {
	assert.Implements(t, (*text.Editor)(nil), Vim)
}
