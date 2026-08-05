package cmd

import (
	"encoding/json"
	"fmt"
	"log"
)

type UpdateCommand struct {
	ID string `name:"id" required:"" help:"The record to update."`
}

func (command *UpdateCommand) Run(rt *runtime) error {
	initial, err := rt.cfg.App.Read(command.ID)
	if err != nil {
		return fmt.Errorf("read record %q: %w", command.ID, err)
	}
	final, err := initial.EditWith(rt.cfg.Editor)
	if err != nil {
		return fmt.Errorf("run editor: %w", err)
	}
	updated, err := rt.cfg.App.Update(command.ID, final)
	if err != nil {
		return fmt.Errorf("update record: %w", err)
	}
	repr, err := json.MarshalIndent(updated, "", "\t")
	if err != nil {
		return fmt.Errorf("represent updated record %q: %w", updated.ID, err)
	}
	log.Printf("successfully updated record %v: %s", updated.ID, repr)
	return nil
}
