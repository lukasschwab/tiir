package store

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"

	// TODO: clean up this imports mess.
	msqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
)

const (
	testDatabaseName    = "tir"
	testDatabaseAddress = "127.0.0.1"
	testDatabasePort    = 3306
)

func startMySQLServer(t testing.TB) *sql.DB {
	engine := msqle.NewDefault(memory.NewDBProvider(memory.NewDatabase(testDatabaseName)))
	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", testDatabaseAddress, testDatabasePort),
	}
	s, err := server.NewDefaultServer(config, engine)
	if err != nil {
		t.Fatalf("Failed creating MySQL server: %v", err)
	}
	go s.Start()
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Logf("Error closing server: %v", err)
		}
	})

	db, err := sql.Open("mysql", fmt.Sprintf("tcp(%s:%d)/%v?parseTime=true", testDatabaseAddress, testDatabasePort, testDatabaseName))
	if err != nil {
		t.Fatalf("Failed opening DB: %v", err)
	}
	return db
}

func TestUseSQL(t *testing.T) {
	db := startMySQLServer(t)

	s, err := UseSQL(db)
	assert.NoError(t, err, "Shouldn't error initializing SQL store")

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

// func BenchmarkUseSQL(b *testing.B) {
// 	texts := make([]*text.Text, b.N)
// 	for i := range texts {
// 		texts[i] = randomText(b)
// 	}

// 	db := startMySQLServer(b)
// 	s, err := UseSQL(db)
// 	if err != nil {
// 		b.Fatalf("Failed starting DB")
// 	}

// 	benchmarkStore(b, s)
// }

// func BenchmarkUseMemory(b *testing.B) {
// 	s := UseMemory()
// 	benchmarkStore(b, s)
// }

func BenchmarkStore(b *testing.B) {
	// FIXME: have to expose store *generators*.

	mysql, err := UseSQL(startMySQLServer(b))
	if err != nil {
		b.Fatalf("Failed starting DB: %v", err)
	}

	fileDb, err := os.CreateTemp(b.TempDir(), "*.json")
	if err != nil {
		b.Fatalf("Failed creating file: %v", err)
	}
	file, err := useFile(fileDb.Name())
	if err != nil {
		b.Fatalf("Error using temp file as DB: %v", err)
	}

	stores := map[string]Interface{
		"memory": UseMemory(),
		"mysql":  mysql,
		"file":   file,
	}

	for name, store := range stores {
		b.Run(name, func(b *testing.B) {
			benchmarkStore(b, store)
		})
	}
}

func benchmarkStore(b *testing.B, store Interface) {
	log.SetOutput(io.Discard)

	texts := make([]*text.Text, b.N)
	for i := range texts {
		texts[i] = randomText(b)
	}

	b.ResetTimer()
	for _, text := range texts {
		store.Upsert(text)
	}
}
