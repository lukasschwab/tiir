package cmd

import (
	"io"
	"log"
	"os"

	"github.com/lukasschwab/tiir/pkg/edit"
	"github.com/lukasschwab/tiir/pkg/tir"
	"github.com/spf13/cobra"
)

// Config properties initialized and closed by rootCmd pre- and post-run funcs.
var (
	configuredService *tir.Service
	configuredEditor  edit.Editor
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tir",
	Short: "Log what you read",
	Long:  `tir ('Today I Read...') is a tool for logging the articles you read.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !verbose {
			log.SetOutput(io.Discard)
		}
		configuredService, _ = tir.FromConfig()
		configuredEditor = edit.Tea{}
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ccli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
}
