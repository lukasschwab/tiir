package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/lukasschwab/tiir/pkg/web"
)

type CreateCommand struct {
	URLs []string `arg:"" optional:"" name:"url" help:"URLs to record."`
}

func (command *CreateCommand) Run(rt *runtime) error {
	if len(command.URLs) == 0 {
		return createFrom(rt, &text.Text{})
	}
	for _, url := range command.URLs {
		if err := createFromURL(rt, url); err != nil {
			return err
		}
	}
	return nil
}

func createFromURL(rt *runtime, url string) error {
	initial, err := web.WebMetadata(url)
	if err != nil {
		log.Printf("couldn't read %q; continuing without metadata: %v", url, err)
		initial = &text.Text{URL: url}
	}
	return createFrom(rt, initial)
}

func createFrom(rt *runtime, initial *text.Text) error {
	final, err := initial.EditWith(rt.cfg.Editor)
	if err != nil {
		return fmt.Errorf("run editor: %w", err)
	}
	created, err := rt.cfg.App.Create(final)
	if err != nil {
		return fmt.Errorf("create record: %w", err)
	}
	repr, err := json.MarshalIndent(created, "", "\t")
	if err != nil {
		return fmt.Errorf("represent created record %q: %w", created.ID, err)
	}
	log.Printf("successfully created record %v: %s", created.ID, repr)
	return nil
}
