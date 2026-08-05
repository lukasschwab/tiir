package cmd

import (
	"encoding/json"
	"fmt"
	"log"
)

type DeleteCommand struct {
	ID string `name:"id" required:"" help:"The record to delete."`
}

func (command *DeleteCommand) Run(rt *runtime) error {
	deleted, err := rt.cfg.App.Delete(command.ID)
	if err != nil {
		return fmt.Errorf("delete record: %w", err)
	}
	repr, err := json.MarshalIndent(deleted, "", "\t")
	if err != nil {
		return fmt.Errorf("represent deleted record %q: %w", deleted.ID, err)
	}
	log.Printf("successfully deleted record %v: %s", deleted.ID, repr)
	return nil
}
