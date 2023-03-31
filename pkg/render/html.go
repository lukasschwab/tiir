package render

import (
	_ "embed" // Compile-time dependency.
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/lukasschwab/tiir/pkg/text"
)

//go:embed templates/html.tmpl
var htmlTemplate string

// HTML table rednering for texts, grouped by their Timestamp date. HTML assumes
// texts are sorted by Timestamp, Descending; see [text.Sort].
func HTML(texts []*text.Text, to io.Writer) error {
	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}
	if err := tmpl.Execute(to, byDate(texts)); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}
	return nil
}

type dateGroup struct {
	Key   string
	Texts []*text.Text
}

func byDate(texts []*text.Text) []*dateGroup {
	aggregate := []*dateGroup{}

	key := func(t time.Time) string {
		return t.Format("January 2, 2006")
	}

	// Assumes texts are sorted by Timestamp descending.
	var currentGroup *dateGroup
	for _, t := range texts {
		if k := key(t.Timestamp); currentGroup != nil && k == currentGroup.Key {
			currentGroup.Texts = append(currentGroup.Texts, t)
		} else {
			currentGroup = &dateGroup{Key: k, Texts: []*text.Text{t}}
			aggregate = append(aggregate, currentGroup)
		}
	}
	return aggregate
}
