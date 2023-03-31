package cmd

import (
	"encoding/json"
	"log"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update your record of a text you read",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if initial, err := configuredService.Read(specifiedTextID); err != nil {
			log.Fatalf("text not found for ID: '%v'", specifiedTextID)
		} else if final, err := configuredEditor.Update(initial); err != nil {
			log.Fatalf("couldn't run editor: %v", err)
		} else if updated, err := configuredService.Update(specifiedTextID, final); err != nil {
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
