package cmd

import (
	"fmt"
)

// ConfigCommand prints the resolved configuration with secrets masked.
type ConfigCommand struct{}

func (command *ConfigCommand) Run(rt *runtime) error {
	contents, err := rt.cfg.MaskedJSON()
	if err != nil {
		return fmt.Errorf("marshal configuration: %w", err)
	}
	_, err = fmt.Fprintln(rt.stdout, string(contents))
	return err
}
