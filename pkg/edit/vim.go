package edit

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/lukasschwab/tiir/pkg/text"
)

// TODO: can we support the user's preferred editor without launching an
// arbitrary application?

// Vim-based [text.Editor]. Uses a temporary file for every Update call.
const Vim vimEditor = iota

type vimEditor int

// Update implements Editor.
func (v vimEditor) Update(initial *text.Text) (final *text.Text, err error) {
	f, err := os.CreateTemp("", "meta.*.json")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(f.Name())

	bytes, err := json.MarshalIndent(editable(*initial), "", "\t")
	if err != nil {
		return nil, fmt.Errorf("error marshaling stored text: %w", err)
	} else if err := os.WriteFile(f.Name(), bytes, os.ModeAppend); err != nil {
		return nil, fmt.Errorf("error initializing temp file: %w", err)
	}

	// Run vim.
	cmd := exec.Command("vim", f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error running editor: %w", err)
	}

	bytes, err = os.ReadFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("error reading temp file: %w", err)
	}

	final = new(text.Text)
	if err := json.Unmarshal(bytes, final); err != nil {
		return nil, fmt.Errorf("couldn't parse user JSON: %w", err)
	}
	return final, nil
}

type editable text.Text

// MarshalJSON marshals the editable subset of e.Text.
//
// NOTE: these must correspond to the JSON tags in text.Text. Any divergence
// may break this editor.
func (e editable) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Title  string `json:"title"`
		URL    string `json:"url"`
		Author string `json:"author"`
		Note   string `json:"note"`
	}{
		Title:  e.Title,
		URL:    e.URL,
		Author: e.Author,
		Note:   e.Note,
	})
}
