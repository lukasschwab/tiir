/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lukasschwab/tiir/pkg/text"
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
		texts, err := configuredService.List()
		if err != nil {
			log.Fatalf("Error listing texts: %v", err)
		}

		// NOTE: this is plaintext rendering.
		// if err := render.Plain.Render(texts, cmd.OutOrStdout()); err != nil {
		// 	log.Fatalf("Error writing texts: %v", err)
		// }

		m := model{list: list.New(items(texts), list.NewDefaultDelegate(), 0, 0)}
		m.list.Title = "Articles"
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatalf("Error running program: %v", err)
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

// tea interface for listing/filtering texts. This is a little awkward: the List
// interface lets us pick two strings to display, but we really have 4-5.

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func items(ts []*text.Text) []list.Item {
	items := make([]list.Item, len(ts))
	for i, t := range ts {
		items[i] = item{t}
	}
	return items
}

type item struct {
	*text.Text
}

func (i item) Title() string       { return fmt.Sprintf("[%v] %v", i.Text.ID, i.Text.Title) }
func (i item) Description() string { return i.Text.Note }
func (i item) FilterValue() string { return i.Title() }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}
