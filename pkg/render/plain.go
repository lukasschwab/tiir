package render

import (
	_ "embed" // Compile-time dependency.
	"fmt"
	"io"
	"text/template"

	"github.com/lukasschwab/tiir/pkg/text"
)

//go:embed templates/plain.tmpl
var plainTextTemplate string

// Plain text rendering for texts.
func Plain(texts []*text.Text, to io.Writer) error {
	tmpl, err := template.New("plain").Parse(plainTextTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}
	if err := tmpl.Execute(to, texts); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}
	return nil
}
