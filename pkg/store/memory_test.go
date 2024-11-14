package store

import (
	"testing"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"
)

func TestUseMemory(t *testing.T) {
	assert.Implements(t, (*Interface)(nil), &Memory{})

	m := UseMemory()
	someText := &text.Text{ID: "some-id"}
	created, err := m.Upsert(someText)
	assert.NoError(t, err, "stores don't do validation")
	assert.Equal(t, someText, created)
}

func TestMemory_Public(t *testing.T) {
	s := UseMemory()

	publicText := randomText(t)
	publicText.Public = true

	privateText := randomText(t)
	privateText.Public = false

	for _, input := range []*text.Text{publicText, privateText} {
		result, err := s.Upsert(input)
		assert.NoError(t, err)
		assert.Equal(t, input.Public, result.Public)

		read, err := s.Read(input.ID)
		assert.NoError(t, err)
		assert.Equal(t, input.Public, read.Public)
	}
}
