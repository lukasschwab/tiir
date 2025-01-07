package cmd

import (
	"encoding/json"
	"log"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/lukasschwab/tiir/pkg/web"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [url]...",
	Short: "Record a text you read",
	Long: `Create a tir record in the configured store. For documentation of store and editor options, see
tir --help.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			from(&text.Text{})
			return
		}
		for _, url := range args {
			fromUrl(url)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func fromUrl(url string) {
	if initial, err := web.WebMetadata(url); err != nil {
		log.Printf("coultn't read '%s'; skipping: %v", url, err)
		from(&text.Text{URL: url})
	} else {
		from(initial)
	}
}

func from(initial *text.Text) {
	if final, err := initial.EditWith(cfg.Editor); err != nil {
		log.Fatalf("couldn't run editor: %v", err)
	} else if created, err := cfg.App.Create(final); err != nil {
		log.Fatalf("error comitting new record: %v", err)
	} else if repr, err := json.MarshalIndent(created, "", "\t"); err != nil {
		log.Fatalf("error representing created record '%v': %v", created.ID, err)
	} else {
		log.Printf("successfully created record %v: %s", created.ID, repr)
	}
}
