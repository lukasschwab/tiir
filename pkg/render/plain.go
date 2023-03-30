package render

import (
	"fmt"
	"io"
	"text/template"

	"github.com/lukasschwab/tiir/pkg/text"
)

type plain struct{}

func (p plain) Render(texts []*text.Text, to io.Writer) error {
	tmpl, err := template.New("plain").ParseFiles("./templates/plain")
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}
	if err := tmpl.Execute(to, texts); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}
	return nil
}
