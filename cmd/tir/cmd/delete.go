package cmd

import (
	"encoding/json"
	"log"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete your record of a text you read",
	Long: `Delete a tir record by ID in the configured store. For store and editor options,
see tir --help.`,
	Run: func(cmd *cobra.Command, args []string) {
		if deleted, err := cfg.App.Delete(specifiedTextID); err != nil {
			log.Fatalf("error deleting record: %v", err)
		} else if repr, err := json.MarshalIndent(deleted, "", "\t"); err != nil {
			log.Fatalf("error representing deleted record '%v': %v", deleted.ID, err)
		} else {
			log.Printf("successfully deleted record %v: %s", deleted.ID, repr)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	requireID(deleteCmd)
}
