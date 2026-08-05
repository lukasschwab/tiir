package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lukasschwab/tiir/pkg/render"
	"github.com/lukasschwab/tiir/pkg/text"
)

var (
	idStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type ListCommand struct {
	Output string `short:"o" enum:"tea,plain,json,jsonfeed,html" default:"tea" help:"Output format for listed texts (tea, plain, json, jsonfeed, html)."`
}

func (command *ListCommand) Run(rt *runtime) error {
	renderFunc, ok := outputRenderers[outputFormat(command.Output)]
	if !ok {
		return invalidOption("output format", command.Output, rendererOptions)
	}

	texts, err := rt.cfg.App.List()
	if err != nil {
		return fmt.Errorf("list texts: %w", err)
	}

	selectedText, err := renderFunc(texts, rt.stdout)
	if err != nil {
		return fmt.Errorf("render texts: %w", err)
	}
	if selectedText == nil {
		return nil
	}
	bytes, err := json.MarshalIndent(selectedText, "", "\t")
	if err != nil {
		return fmt.Errorf("represent selected record: %w", err)
	}
	_, err = fmt.Fprintln(rt.stdout, string(bytes))
	return err
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

type renderFunc func(texts []*text.Text, output io.Writer) (selected *text.Text, err error)

// cli adapter for render.Functions.
func cli(f render.Function) renderFunc {
	return func(texts []*text.Text, output io.Writer) (*text.Text, error) {
		return nil, f(texts, output)
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
func renderTea(texts []*text.Text, _ io.Writer) (*text.Text, error) {
	m := model{list: list.New(items(texts), list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "Articles"
	m.list.Filter = list.UnsortedFilter
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

func (i item) Title() string {
	date := idStyle.Render(i.Text.Timestamp.Format("06-01-02"))
	return fmt.Sprintf("%s %v", date, i.Text.Title)
}

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
