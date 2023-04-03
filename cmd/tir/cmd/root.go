package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/lukasschwab/tiir/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Config initialized and closed by rootCmd pre- and post-run funcs.
	cfg *config.Config

	// storeOptions group the available StoreTypes for rendering CLI helper
	// text; it matches the storeFactories map keyset.
	storeOptions = []string{
		string(config.StoreTypeFile),
		string(config.StoreTypeMemory),
		string(config.StoreTypeHTTP),
	}

	// editorOptions group the available EditorTypes for rendering CLI helper
	// text; it matches the editors map keyset.
	editorOptions = []string{
		string(config.EditorTypeVim),
		string(config.EditorTypeTea),
	}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tir",
	Short: "Log what you read",
	Long:  `tir ('Today I Read...') is a tool for logging the articles you read.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if !verbose {
			log.SetOutput(io.Discard)
		}

		if cfg, err = config.Load(); err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return cfg.App.Close()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")

	flagStore := "store"
	rootCmd.PersistentFlags().StringP(flagStore, "s", "file", fmt.Sprintf("store to use (%v)", strings.Join(storeOptions, ", ")))
	viper.BindPFlag(config.KeyStoreType, rootCmd.PersistentFlags().Lookup(flagStore))

	flagFileLocation := "file-location"
	rootCmd.PersistentFlags().String(flagFileLocation, "$HOME/.tir.json", "if store is 'file,' specifies file to use")
	viper.BindPFlag(config.KeyFileStoreLocation, rootCmd.PersistentFlags().Lookup(flagFileLocation))

	flagBaseURL := "base-url"
	rootCmd.PersistentFlags().String(flagBaseURL, "", "when store is 'http,' specifies service URL to use")
	viper.BindPFlag(config.KeyHTTPStoreBaseURL, rootCmd.PersistentFlags().Lookup(flagBaseURL))
	flagAPISecret := "api-secret"
	rootCmd.PersistentFlags().String(flagAPISecret, "", "when store is 'http,' specifies API secret to authorize requests")
	viper.BindPFlag(config.KeyHTTPStoreAPISecret, rootCmd.PersistentFlags().Lookup(flagAPISecret))

	flagEditor := "editor"
	rootCmd.PersistentFlags().StringP(flagEditor, "e", "tea", fmt.Sprintf("editor to use (%v)", strings.Join(editorOptions, ", ")))
	viper.BindPFlag(config.KeyEditor, rootCmd.PersistentFlags().Lookup(flagEditor))
}
