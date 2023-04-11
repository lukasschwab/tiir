package render

import (
	_ "embed" // Compile-time dependency.
	"fmt"
	"html/template"
	"io"
	"time"

	"github.com/lukasschwab/tiir/pkg/text"
)

//go:embed templates/html.tmpl
var htmlTemplate string

// HTML table rendering for texts, grouped by [text.Text.Timestamp]. HTML
// assumes texts it receives are already ordered by timestamp, descending; see
// [text.Sort].
//
// Example output:
//
//	<head>
//		<!-- Head contents... -->
//	</head>
//	<h1 class="title">tir</h1>
//	<p><a href="https://github.com/lukasschwab/tiir">GitHub</a></p>
//	<hr/>
//	<table class="table">
//		<tr>
//			<th>Title</th>
//			<th>Author</th>
//			<th>Note</th>
//			<th>Date</th>
//		</tr>
//		<td colspan="4"><h3>April 7, 2023</h3></td>
//		<tr>
//			<td><a href="https://davidchall.github.io/ggip/articles/visualizing-ip-data.html">Visualizing IP data</a></td>
//			<td>David Hall</td>
//			<td>Use a Hilbert Curve: efficient 2D packing that keeps consecutive sequences spatially contiguous.</td>
//			<td>2023-04-07</td>
//		</tr>
//		<!-- More rows... -->
//	</table>
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
