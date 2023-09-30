package store

import (
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"
	// TODO: clean up this imports mess.
)

func TestUseSQL(t *testing.T) {
	s := BuildMySQL(t)

	// Initial.
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
}

func randomText(t testing.TB) *text.Text {
	id, err := text.RandomID()
	if err != nil {
		t.Fatalf("Failed generating random ID: %v", err)
	}
	return &text.Text{
		Title:  "My Text",
		URL:    "https://www.google.com",
		Author: "Enkidu Gilgamesh",
		Note:   "This is a test text",
		ID:     id,
		// Truncate to elide lower precision in MySQL.
		Timestamp: time.Now().UTC().Truncate(time.Second),
	}
}

// func BenchmarkStore(b *testing.B) {
// 	// FIXME: have to expose store *generators*.

// 	// mysql, err := UseSQL(startMySQLServer(b))
// 	mysql := BuildMySQL(b)

// 	fileDb, err := os.CreateTemp(b.TempDir(), "*.json")
// 	if err != nil {
// 		b.Fatalf("Failed creating file: %v", err)
// 	}
// 	file, err := useFile(fileDb.Name())
// 	if err != nil {
// 		b.Fatalf("Error using temp file as DB: %v", err)
// 	}

// 	stores := map[string]Interface{
// 		"memory": UseMemory(),
// 		"mysql":  mysql,
// 		"file":   file,
// 	}

// 	for name, store := range stores {
// 		b.Run(name, func(b *testing.B) {
// 			benchmarkStore(b, store)
// 		})
// 	}
// }

// func benchmarkStore(b *testing.B, store Interface) {
// 	log.SetOutput(io.Discard)

// texts := make([]*text.Text, b.N)
// for i := range texts {
// 	texts[i] = randomText(b)
// }

// b.ResetTimer()
// for _, text := range texts {
// 	store.Upsert(text)
// }
// }
