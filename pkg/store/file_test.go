package store

import (
	"os"
	"testing"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"
)

func TestUseFile(t *testing.T) {
	assert.Implements(t, (*Store)(nil), &file{})

	f, err := os.CreateTemp(t.TempDir(), "*.json")
	assert.NoError(t, err)
	assert.NoError(t, f.Close())

	fStore, err := useFile(f.Name())
	assert.NoError(t, err)

	inner, err := fStore.parse()
	assert.NoError(t, err)
	assert.Empty(t, inner.texts)

	m := useMemory(&text.Text{ID: "abc123de"})

	assert.NoError(t, fStore.commit(m))

	inner, err = fStore.parse()
	assert.NoError(t, err)
	assert.Contains(t, inner.texts, "abc123de")

	assert.NoError(t, fStore.Close())

	fStore, err = useFile(f.Name())
	assert.NoError(t, err, "can reopen previously-opened file")
	defer fStore.Close()

	inner, err = fStore.parse()
	assert.NoError(t, err)
	assert.Contains(t, inner.texts, "abc123de", "records should persist when file is closed")
}
