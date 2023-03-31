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
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		initial := &text.Text{}
		// TODO: parameterize the editor.
		if final, err := configEditor().Update(initial); err != nil {
			log.Fatalf("couldn't run editor: %v", err)
		} else if created, err := configService().Create(final); err != nil {
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
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
