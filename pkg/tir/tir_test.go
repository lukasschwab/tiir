package tir

import (
	"io"
	"testing"

	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	assert.Implements(t, (*io.Closer)(nil), new(app))

	s := New(store.UseMemory(), false)

	original := &text.Text{Author: "a", Note: "n", URL: "u", Title: "t"}

	created, err := s.Create(original)
	assert.NoError(t, err)
	assert.Equal(t, created.Author, original.Author)
	assert.Equal(t, created.Note, original.Note)
	assert.Equal(t, created.URL, original.URL)
	assert.NotEmpty(t, created.ID, "should create ID before creating text")

	// No-op update.
	updated, err := s.Update(created.ID, &text.Text{})
	assert.NoError(t, err)
	assert.Equal(t, created, updated)

	updated, err = s.Update(created.ID, &text.Text{Author: "New Author"})
	assert.NoError(t, err)
	assert.Equal(t, "New Author", updated.Author)
	reRead, err := s.Read(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "New Author", reRead.Author)

	deleted, err := s.Delete(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, deleted, reRead)

	_, err = s.Read(created.ID)
	assert.Error(t, err)
}

func TestList_NonpublicText(t *testing.T) {
	public, private := New(store.UseMemory(), true), New(store.UseMemory(), false)

	for _, s := range []Interface{public, private} {
		_, err := s.Create(&text.Text{Author: "a", Note: "n", URL: "u", Title: "t", Public: false})
		assert.NoError(t, err)
	}

	publicTexts, err := public.List()
	assert.NoError(t, err)
	assert.Len(t, publicTexts, 0)

	allTexts, err := private.List()
	assert.NoError(t, err)
	assert.Len(t, allTexts, 1)
}

func TestList_PublicText(t *testing.T) {
	public, private := New(store.UseMemory(), true), New(store.UseMemory(), false)

	for _, s := range []Interface{public, private} {
		_, err := s.Create(&text.Text{Author: "a", Note: "n", URL: "u", Title: "t", Public: true})
		assert.NoError(t, err)

		texts, err := s.List()
		assert.NoError(t, err)
		assert.Len(t, texts, 1)
	}
}

func TestValidation(t *testing.T) {
	s := New(store.UseMemory(), false)

	created, err := s.Create(&text.Text{})
	assert.Error(t, err)
	assert.Nil(t, created)
}
