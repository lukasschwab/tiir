package store

import (
	"os"
	"testing"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"
)

func TestUseFile(t *testing.T) {
	assert.Implements(t, (*Interface)(nil), &file{})

	db, err := os.CreateTemp(t.TempDir(), "*.json")
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	// Initial store: from empty DB.
	f, err := useFile(db.Name())
	assert.NoError(t, err)

	err = f.load()
	assert.NoError(t, err)
	assert.Empty(t, f.cache.texts)

	f.cache = useMemory(&text.Text{ID: "abc123de"})

	assert.NoError(t, f.commit())

	err = f.load()
	assert.NoError(t, err)
	assert.Contains(t, f.cache.texts, "abc123de")

	assert.NoError(t, f.Close())

	// Second store.
	f2, err := useFile(db.Name())
	assert.NoError(t, err, "can reopen previously-opened file")
	defer f2.Close()

	err = f2.load()
	assert.NoError(t, err)
	assert.Contains(t, f2.cache.texts, "abc123de", "records should persist when file is closed")
}
