package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Flags.
var (
	specifiedTextID string
	verbose         bool
)

// requireID requires a standard ID parameter identifying an extant record.
func requireID(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&specifiedTextID, "id", "", fmt.Sprintf("The record to %v.", cmd.Name()))
	cmd.MarkPersistentFlagRequired("id")
}
