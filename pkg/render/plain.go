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

// Plain text rendering for texts. Example output:
//
//	[35bb8126] Visualizing IP data (2023-4-7)
//	David Hall @ https://davidchall.github.io/ggip/articles/visualizing-ip-data.html
//
//		Use a Hilbert Curve: efficient 2D packing that keeps consecutive sequences spatially contiguous.
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
