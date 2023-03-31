package edit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/lukasschwab/tiir/pkg/text"
)

// TODO: can we support the user's preferred editor without launching an
// arbitrary application?

// Vim editor for texts. Uses a temporary file for every Update call.
type Vim struct{}

type editable struct {
	Title  string
	URL    string
	Author string
	Note   string
}

func pick(t *text.Text) editable {
	return editable{
		Title:  t.Title,
		URL:    t.URL,
		Author: t.Author,
		Note:   t.Note,
	}
}

// Update implements Editor.
func (v Vim) Update(initial *text.Text) (final *text.Text, err error) {
	f, err := os.CreateTemp("", "meta.*.json")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(f.Name())

	// TODO: filter down to just the editable fields.
	bytes, err := json.MarshalIndent(pick(initial), "", "\t")
	if err != nil {
		return nil, fmt.Errorf("error marshaling stored text: %w", err)
	} else if err := ioutil.WriteFile(f.Name(), bytes, os.ModeAppend); err != nil {
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

	bytes, err = ioutil.ReadFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("error reading temp file: %w", err)
	}

	final = new(text.Text)
	if err := json.Unmarshal(bytes, final); err != nil {
		return nil, fmt.Errorf("couldn't parse user JSON: %w", err)
	}
	return final, nil
}
