package render

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"

	"github.com/lukasschwab/tiir/pkg/text"
)

//go:embed templates/html
var htmlTemplate string

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
