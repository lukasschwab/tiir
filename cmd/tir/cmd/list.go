/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/lukasschwab/tiir/pkg/render"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if texts, err := configuredService.List(); err != nil {
			log.Fatalf("Error listing texts: %v", err)
		} else if err := render.Plain.Render(texts, cmd.OutOrStdout()); err != nil {
			log.Fatalf("Error writing texts: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// TODO: parameterize the renderer, and perhaps also the sort order.

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
