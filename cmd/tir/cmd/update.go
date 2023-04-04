package cmd

import (
	"encoding/json"
	"log"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"edit"},
	Short:   "Update your record of a text you read",
	Long: `Update a tir record by ID in the configured store. For store and editor options,
see tir --help.`,
	Run: func(cmd *cobra.Command, args []string) {
		if initial, err := cfg.App.Read(specifiedTextID); err != nil {
			log.Fatalf("text not found for ID: '%v'", specifiedTextID)
		} else if final, err := initial.EditWith(cfg.Editor); err != nil {
			log.Fatalf("couldn't run editor: %v", err)
		} else if updated, err := cfg.App.Update(specifiedTextID, final); err != nil {
			log.Fatalf("error comitting new record: %v", err)
		} else if repr, err := json.MarshalIndent(updated, "", "\t"); err != nil {
			log.Fatalf("error representing updated record '%v': %v", updated.ID, err)
		} else {
			log.Printf("successfully updated record %v: %s", updated.ID, repr)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	requireID(updateCmd)
}
