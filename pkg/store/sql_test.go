package store

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"

	_ "github.com/libsql/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

func startLocalLibSQL(t testing.TB) *SQL {
	// Initialize an empty DB. In future tests, could initiate a non-empty one
	// for testing (e.g. with existing table, entries).
	emptyFile, err := os.CreateTemp(t.TempDir(), "*.db")
	assert.NoError(t, err)
	t.Logf("Using DB at file %v", emptyFile.Name())

	connectionString := fmt.Sprintf("file://%s", emptyFile.Name())
	s, err := useLibSQL(connectionString)
	if err != nil {
		t.Fatalf("Failed initializing LibSQL store: %v", err)
	}
	return s
}

func TestUseLibSQL(t *testing.T) {
	// Initialize.
	s := startLocalLibSQL(t)
	texts, err := s.List(text.Timestamps, text.Descending)
	assert.NoError(t, err, "Shouldn't error listing on empty database")
	assert.Empty(t, texts)

	// Insert.
	firstText := randomText(t)
	upserted, err := s.Upsert(firstText)
	assert.NoError(t, err)
	assert.Equal(t, firstText, upserted)

	read, err := s.Read(firstText.ID)
	assert.NoError(t, err)
	assert.Equal(t, firstText, read)

	// Update.
	firstText.Author = firstText.Author + " Jr."
	upserted, err = s.Upsert(firstText)
	assert.NoError(t, err)
	assert.Equal(t, firstText, upserted)

	texts, err = s.List(text.Timestamps, text.Descending)
	assert.NoError(t, err)
	assert.Len(t, texts, 1)

	// Second insert.
	secondText := randomText(t)
	secondText.Timestamp = secondText.Timestamp.Add(1 * time.Second)
	upserted, err = s.Upsert(secondText)
	assert.NoError(t, err)
	assert.Equal(t, secondText, upserted)

	texts, err = s.List(text.Timestamps, text.Descending)
	assert.NoError(t, err)
	assert.Len(t, texts, 2)
	assert.Equal(t, secondText.ID, texts[0].ID)
	assert.Equal(t, firstText.ID, texts[1].ID)

	// Delete.
	deleted, err := s.Delete(secondText.ID)
	assert.NoError(t, err)
	assert.Equal(t, secondText, deleted)

	texts, err = s.List(text.Timestamps, text.Descending)
	assert.NoError(t, err)
	assert.Len(t, texts, 1)
	assert.Equal(t, firstText, texts[0])
}

func TestSQL_Public(t *testing.T) {
	s := startLocalLibSQL(t)

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

func randomText(t testing.TB) *text.Text {
	id, err := text.RandomID()
	if err != nil {
		t.Fatalf("Failed generating random ID: %v", err)
	}
	return &text.Text{
		Title:     "My Text",
		URL:       "https://www.google.com",
		Author:    "Enkidu Gilgamesh",
		Note:      "This is a test text",
		ID:        id,
		Timestamp: time.Now().UTC(),
	}
}
