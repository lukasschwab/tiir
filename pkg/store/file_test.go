package store

import (
	"os"
	"testing"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"
)

func TestUseFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.json")
	assert.NoError(t, err)
	assert.NoError(t, f.Close())

	file, err := UseFile(f.Name())
	assert.NoError(t, err)

	inner, err := file.parse()
	assert.NoError(t, err)
	assert.Empty(t, inner.texts)

	m := UseMemory(&text.Text{ID: "abc123de"})

	assert.NoError(t, file.commit(m))

	inner, err = file.parse()
	assert.NoError(t, err)
	assert.Contains(t, inner.texts, "abc123de")

	assert.NoError(t, file.Close())

	file, err = UseFile(f.Name())
	assert.NoError(t, err, "can reopen previously-opened file")
	defer file.Close()

	inner, err = file.parse()
	assert.NoError(t, err)
	assert.Contains(t, inner.texts, "abc123de", "records should persist when file is closed")
}
