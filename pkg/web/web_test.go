package web_test

import (
	"testing"

	"github.com/lukasschwab/tiir/pkg/web"
	"github.com/stretchr/testify/assert"
)

const (
	// FIXME: brittle. Probably best to replace with some golden files, e.g.
	// strings in some separate go package.
	testSite = "https://www.geoffreylitt.com/2024/12/22/making-programming-more-fun-with-an-ai-generated-debugger.html"
)

func TestMetadata(t *testing.T) {
	initial, err := web.Metadata(testSite)

	assert.NoError(t, err)
	assert.Equal(t, "AI-generated tools can make programming more fun", initial.Title)
	assert.Equal(t, "", initial.Author)
}
