package store

import (
	"testing"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"
)

func TestNewMemory(t *testing.T) {
	m := UseMemory()
	someText := &text.Text{ID: "some-id"}
	created, err := m.Upsert(someText)
	assert.NoError(t, err, "stores don't do validation")
	assert.Equal(t, someText, created)
	assert.Implements(t, (*Store)(nil), m)
}
