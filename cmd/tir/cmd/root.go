package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/lukasschwab/tiir/pkg/tir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config properties initialized and closed by rootCmd pre- and post-run funcs.
var (
	configuredService *tir.Service
	configuredEditor  text.Editor
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
		if configuredService, configuredEditor, err = tir.FromConfig(); err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return configuredService.Close()
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
	rootCmd.PersistentFlags().StringP(flagStore, "s", "file", fmt.Sprintf("store to use (%v)", strings.Join(tir.StoreOptions, ", ")))
	viper.BindPFlag(tir.ConfigStoreType, rootCmd.PersistentFlags().Lookup(flagStore))

	flagFileLocation := "file-location"
	rootCmd.PersistentFlags().String(flagFileLocation, "$HOME/.tir.json", "if store is 'file,' specifies file to use")
	viper.BindPFlag(tir.ConfigFileStoreLocation, rootCmd.PersistentFlags().Lookup(flagFileLocation))

	flagBaseURL := "base-url"
	rootCmd.PersistentFlags().String(flagBaseURL, "", "when store is 'http,' specifies service URL to use")
	viper.BindPFlag(tir.ConfigHTTPStoreBaseURL, rootCmd.PersistentFlags().Lookup(flagBaseURL))

	flagEditor := "editor"
	rootCmd.PersistentFlags().StringP(flagEditor, "e", "tea", fmt.Sprintf("editor to use (%v)", strings.Join(tir.EditorOptions, ", ")))
	viper.BindPFlag(tir.ConfigEditor, rootCmd.PersistentFlags().Lookup(flagEditor))
}
