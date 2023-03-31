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
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		service := configService()

		if deleted, err := service.Delete(specifiedTextID); err != nil {
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
