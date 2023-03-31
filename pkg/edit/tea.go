package edit

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lukasschwab/tiir/pkg/text"
)

// Tea based text.Editor for the command line..
const Tea teaEditor = iota

type teaEditor int

// Update implements Editor.
func (t teaEditor) Update(initial *text.Text) (final *text.Text, err error) {
	final = new(text.Text)
	if _, err = tea.NewProgram(initialModel(initial, final)).Run(); err != nil {
		err = fmt.Errorf("could not start tea editor: %w", err)
	}
	return final, err
}

// Styles.
var (
	focusedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	completedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("27"))
	noStyle        = lipgloss.NewStyle()

	buttonText         = ":Create new record"
	focusedButtonStyle = lipgloss.NewStyle().Background(lipgloss.Color("205"))
	focusedButton      = focusedButtonStyle.Render(buttonText)
	blurredButton      = blurredStyle.Render(buttonText)
)

const (
	urlInputIndex    = 0
	titleInputIndex  = 1
	authorInputIndex = 2
	noteInputIndex   = 3
)

type model struct {
	result *text.Text

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

// commit the edited state to the model result.
func (m model) commit() {
	*m.result = *m.toText()
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
	fmt.Fprintf(&b, "\n%s\n", *button)

	return b.String()
}

func initialModel(initial, result *text.Text) model {
	m := model{result: result}

	for index, data := range map[int][2]string{
		urlInputIndex:    {"   URL> ", initial.URL},
		titleInputIndex:  {" Title> ", initial.Title},
		authorInputIndex: {"Author> ", initial.Author},
		noteInputIndex:   {"  Note> ", initial.Note},
	} {
		input := textinput.New()
		// NOTE: would it be cleaner to indicate fields using 'Placeholder?'
		// Probably cleaner but less usable.
		input.Prompt = data[0]
		input.SetValue(data[1])
		m.inputs[index] = input
	}

	// Focus the URL input initially.
	m.inputs[urlInputIndex].Focus()
	m.inputs[urlInputIndex].PromptStyle = focusedStyle
	m.inputs[urlInputIndex].TextStyle = focusedStyle

	return m
}
