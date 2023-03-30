package render

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/stretchr/testify/assert"
)

func TestRenderPlainText(t *testing.T) {
	texts := make([]*text.Text, 5)
	for i := range texts {
		someText := &text.Text{
			ID:     fmt.Sprintf("adadada%d", i),
			Title:  fmt.Sprintf("My %dth Text", i),
			URL:    fmt.Sprintf("github.com/lukasschwab/tiir/pkg/text/%d", i),
			Author: "L. Schwab",
			Note:   fmt.Sprintf("This is my note for text #%d. What a good text!", i),
		}
		assert.NoError(t, someText.Validate())
		texts[i] = someText
	}

	var rendered strings.Builder
	err := Plain.Render(texts, &rendered)
	assert.NoError(t, err)

	t.Logf(rendered.String())
}
