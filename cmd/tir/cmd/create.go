package cmd

import (
	"encoding/json"
	"log"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Record a text you read",
	Long: `Create a tir record in the configured store. For store and editor options, see
tir --help.`,
	Run: func(cmd *cobra.Command, args []string) {
		initial := &text.Text{Public: true}
		if final, err := initial.EditWith(cfg.Editor); err != nil {
			log.Fatalf("couldn't run editor: %v", err)
		} else if created, err := cfg.App.Create(final); err != nil {
			log.Fatalf("error comitting new record: %v", err)
		} else if repr, err := json.MarshalIndent(created, "", "\t"); err != nil {
			log.Fatalf("error representing created record '%v': %v", created.ID, err)
		} else {
			log.Printf("successfully created record %v: %s", created.ID, repr)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
