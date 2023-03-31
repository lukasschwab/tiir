package render

import (
	_ "embed"
	"fmt"
	"io"
	"text/template"

	"github.com/lukasschwab/tiir/pkg/text"
)

type plain struct{}

//go:embed templates/plain
var plainTextTemplate string

func (p plain) Render(texts []*text.Text, to io.Writer) error {
	tmpl, err := template.New("plain").Parse(plainTextTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}
	if err := tmpl.Execute(to, texts); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}
	return nil
}
