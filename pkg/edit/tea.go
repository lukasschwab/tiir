package edit

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/lukasschwab/tiir/pkg/text"
)

const Tea teaEditor = iota

type teaEditor int

func (h teaEditor) Update(initial *text.Text) (final *text.Text, err error) {
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
			huh.NewConfirm().
				Title("Share publicly?").
				Affirmative("Public").
				Negative("Not public").
				Value(&t.Public),
		),
	)
}
