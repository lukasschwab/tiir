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
		string(config.StoreTypeLibSQL),
	}

	// editorOptions group the available EditorTypes for rendering CLI helper
	// text; it matches the editors map keyset.
	editorOptions = []string{
		string(config.EditorTypeVim),
		string(config.EditorTypeTea),
	}

	// specifiedTextID set for update and delete commands.
	specifiedTextID string

	// verbose set by --verbose flag.
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tir",
	Short: "Log what you read",
	Long: `tir ('Today I Read...') is a tool for logging the articles you read.

By default, it writes a JSON collection to $HOME/.tir.json, with an interactive
CLI for adding new readings.

Store readings elsewhere by specifying an alternate --store:

+ 'file' (default): a local JSON file. Optionally uses --file-location.
+ 'http': a hosted tir service. Requires --base-url; some hosted services will
  also require --api-secret.
+ 'memory': an in-memory store that doesn't persist data between calls.

Specify an editor for creating and updating records:

+ 'tea' (default): interactive CLI.
+ 'vim': open the record in a temporary file in vim.`,
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
	bindPFlag(config.KeyStoreType, flagStore)

	flagFileLocation := "file-location"
	rootCmd.PersistentFlags().String(flagFileLocation, "$HOME/.tir.json", "when store is 'file,' specifies file to use")
	bindPFlag(config.KeyFileStoreLocation, flagFileLocation)

	flagBaseURL := "base-url"
	rootCmd.PersistentFlags().String(flagBaseURL, "", "when store is 'http,' specifies service URL to use")
	bindPFlag(config.KeyHTTPStoreBaseURL, flagBaseURL)
	flagAPISecret := "api-secret"
	rootCmd.PersistentFlags().String(flagAPISecret, "", "when store is 'http,' specifies API secret to authorize requests")
	bindPFlag(config.KeyHTTPStoreAPISecret, flagAPISecret)

	flagConnectionString := "connection-string"
	rootCmd.PersistentFlags().String(flagConnectionString, "", "when store is 'libsql,' specifies where to connect")
	bindPFlag(config.KeyLibSQLStoreConnectionString, flagConnectionString)

	flagEditor := "editor"
	rootCmd.PersistentFlags().StringP(flagEditor, "e", "tea", fmt.Sprintf("editor to use (%v)", strings.Join(editorOptions, ", ")))
	bindPFlag(config.KeyEditor, flagEditor)
}

// bindPFlag in viper specified by configKey to the persistent cobra flag with
// flagName.
func bindPFlag(configKey, flagName string) {
	if err := viper.BindPFlag(
		configKey,
		rootCmd.PersistentFlags().Lookup(flagName),
	); err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}
}

// requireID requires a standard ID parameter identifying an extant record.
func requireID(cmd *cobra.Command) {
	const flagID = "id"
	cmd.PersistentFlags().StringVar(&specifiedTextID, flagID, "", fmt.Sprintf("The record to %v.", cmd.Name()))
	if err := cmd.MarkPersistentFlagRequired(flagID); err != nil {
		log.Fatalf("Error marking %v flag required: %v", flagID, err)
	}
}
