package main

import (
	"fmt"
	"log" // TODO: set up zap.
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/lukasschwab/tiir/pkg/tir"
)

// Styles.
var (
	focusedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	completedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("27"))
	noStyle        = lipgloss.NewStyle()

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

const (
	urlInputIndex    = 0
	titleInputIndex  = 1
	authorInputIndex = 2
	noteInputIndex   = 3
)

type model struct {
	service tir.Service

	focusIndex int
	// URL, Title, Author, Notes.
	inputs [4]textinput.Model
}

func (m model) urlInput() textinput.Model {
	return m.inputs[urlInputIndex]
}

func (m model) titleInput() textinput.Model {
	return m.inputs[titleInputIndex]
}

func (m model) authorInput() textinput.Model {
	return m.inputs[authorInputIndex]
}

func (m model) noteInput() textinput.Model {
	return m.inputs[noteInputIndex]
}

func (m model) toText() *text.Text {
	// NOTE: should we validate here, to prevent premature submission?
	return &text.Text{
		URL:    m.urlInput().Value(),
		Title:  m.titleInput().Value(),
		Author: m.authorInput().Value(),
		Note:   m.noteInput().Value(),
	}
}

func (m model) canSubmit() bool {
	return m.toText().Validate() == nil
}

func (m model) commit() {
	// TODO: handle validation.
	// TODO; display an error text if this returns an error. Check out the old
	// cursor mode pattern for an example.
	t, err := m.service.Create(m.toText())
	if err != nil {
		log.Fatalf("Failed to write text: %v", err)
	} else {
		log.Printf("Successfully wrote text [%s]", t.ID)
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) wrapFocusIndex(proposedIndex int) int {
	// Allow the user to highlight the "submit" button (see m.View)...
	maxIndex := len(m.inputs)
	// ...unless the input is invalid!
	if !m.canSubmit() {
		maxIndex--
	}

	if proposedIndex > maxIndex {
		return 0
	} else if proposedIndex < 0 {
		return maxIndex
	} else {
		return proposedIndex
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			if s == "enter" && m.focusIndex == len(m.inputs) {
				// TODO: validation before "allowing" submission.
				m.commit()
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			m.focusIndex = m.wrapFocusIndex(m.focusIndex)

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				if m.inputs[i].Value() != "" {
					m.inputs[i].PromptStyle = completedStyle
				}
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("Creating a new record...\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}

func initialModel(s tir.Service) model {
	m := model{service: s}

	for index, prompt := range map[int]string{
		urlInputIndex:    "   URL> ",
		titleInputIndex:  " Title> ",
		authorInputIndex: "Author> ",
		noteInputIndex:   "  Note> ",
	} {
		input := textinput.New()
		// NOTE: would it be cleaner to indicate fields using 'Placeholder?'
		// Probably cleaner but less usable.
		input.Prompt = prompt
		m.inputs[index] = input
	}

	// Focus the URL input initially.
	m.inputs[urlInputIndex].Focus()
	m.inputs[urlInputIndex].PromptStyle = focusedStyle
	m.inputs[urlInputIndex].TextStyle = focusedStyle

	return m
}

func main() {
	// TODO: parse command line arguments. This might be easier if we factor the
	// editor interface out from the data process.
	// TODO: actually persist results.
	service := tir.Service{Store: store.NewMemory()}

	if _, err := tea.NewProgram(initialModel(service)).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
