package store

import (
	"testing"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"
)

func TestNewMemory(t *testing.T) {
	m := NewMemory()
	m.Create(&text.Text{ID: "some-id"})
	assert.Implements(t, (*Store)(nil), m)
}
