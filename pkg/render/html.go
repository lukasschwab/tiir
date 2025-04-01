package render

import (
	_ "embed" // Compile-time dependency.
	"fmt"
	"html/template"
	"io"

	"github.com/lukasschwab/tiir/pkg/text"
)

//go:embed templates/html.tmpl
var htmlTemplate string

// HTML table rendering for texts. HTML assumes texts it receives are already
// ordered by timestamp, descending; see [text.Sort].
func HTML(texts []*text.Text, to io.Writer) error {
	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}
	if err := tmpl.Execute(to, texts); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}
	return nil
}
