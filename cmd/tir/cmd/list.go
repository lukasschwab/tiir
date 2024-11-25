package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lukasschwab/tiir/pkg/render"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/spf13/cobra"
)

var output string

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all the texts you recorded reading",
	Long: `List all tir records in the configured store. For store and editor options, see
tir --help.`,
	Run: func(cmd *cobra.Command, args []string) {
		renderFunc, ok := outputRenderers[outputFormat(output)]
		if !ok {
			log.Fatalf("Invalid renderer type '%s'; use one of %v", output, strings.Join(rendererOptions, ", "))
		}

		texts, err := cfg.App.List()
		if err != nil {
			log.Fatalf("Error listing texts: %v", err)
		}

		if selectedText, err := renderFunc(texts, cmd); err != nil {
			log.Fatalf("error rendering texts: %v", err)
		} else if bytes, err := json.MarshalIndent(selectedText, "", "\t"); selectedText != nil && err == nil {
			fmt.Println(string(bytes))
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.PersistentFlags().StringVarP(&output, "output", "o", string(OutputTea), fmt.Sprintf("output format for listed texts (%v)", strings.Join(rendererOptions, ", ")))
}

type outputFormat string

// Output formats the user can select.
const (
	OutputTea      outputFormat = "tea"
	OutputPlain    outputFormat = "plain"
	OutputJSON     outputFormat = "json"
	OutputJSONFeed outputFormat = "jsonfeed"
	OutputHTML     outputFormat = "html"
)

var rendererOptions = []string{
	string(OutputTea),
	string(OutputPlain),
	string(OutputJSON),
	string(OutputJSONFeed),
	string(OutputHTML),
}

type renderFunc func(texts []*text.Text, cmd *cobra.Command) (selected *text.Text, err error)

// cli adapter for render.Functions.
func cli(f render.Function) renderFunc {
	return func(texts []*text.Text, cmd *cobra.Command) (*text.Text, error) {
		return nil, f(texts, cmd.OutOrStdout())
	}
}

// outputRenderers by outputFormat.
var outputRenderers = map[outputFormat]renderFunc{
	OutputTea:      renderTea,
	OutputPlain:    cli(render.Plain),
	OutputJSON:     cli(render.JSON),
	OutputJSONFeed: cli(render.JSONFeed),
	OutputHTML:     cli(render.HTML),
}

// renderTea renders a tea interface for listing/filtering texts. This is a
// little awkward: the List interface lets us pick two strings to display, but
// we really have 4-5.
func renderTea(texts []*text.Text, cmd *cobra.Command) (*text.Text, error) {
	m := model{list: list.New(items(texts), list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Articles"
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	return finalModel.(model).finalSelection.Text, err
}

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
	// TODO: consider allowing multiple selections.
	finalSelection item
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		} else if msg.String() == "enter" {
			m.finalSelection = m.list.SelectedItem().(item)
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
