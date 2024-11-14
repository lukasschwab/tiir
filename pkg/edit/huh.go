package edit

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/lukasschwab/tiir/pkg/text"
)

const Huh huhEditor = iota

type huhEditor int

func (h huhEditor) Update(initial *text.Text) (final *text.Text, err error) {
	if err = form(initial).Run(); err != nil {
		return nil, fmt.Errorf("could not edit text: %w", err)
	}
	return initial, nil
}

func form(t *text.Text) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("URL").
				Value(&t.URL),
			huh.NewInput().
				Title("Title").
				Value(&t.Title),
			huh.NewInput().
				Title("Author").
				Value(&t.Author),
			huh.NewInput().
				Title("Note").
				Value(&t.Note),
		),
	)
}
