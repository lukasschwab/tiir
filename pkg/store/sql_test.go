package store

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"

	// In-memory SQL server for testing.
	sqle "github.com/dolthub/go-mysql-server"
	sqleMemory "github.com/dolthub/go-mysql-server/memory"
	sqleServer "github.com/dolthub/go-mysql-server/server"
)

const (
	testDatabaseName    = "tir"
	testDatabaseAddress = "127.0.0.1"
	testDatabasePort    = 3306
)

func startMySQLServer(t testing.TB) *sql.DB {
	// Launch in-memory server with empty DB.
	engine := sqle.NewDefault(sqleMemory.NewDBProvider(sqleMemory.NewDatabase(testDatabaseName)))
	config := sqleServer.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", testDatabaseAddress, testDatabasePort),
	}
	s, err := sqleServer.NewDefaultServer(config, engine)
	if err != nil {
		t.Fatalf("Failed creating MySQL server: %v", err)
	}
	go s.Start()
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Logf("Error closing server: %v", err)
		}
	})

	// Connect MySQL driver to that DB.
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
