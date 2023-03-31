package cmd

import (
	"github.com/lukasschwab/tiir/pkg/edit"
	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/lukasschwab/tiir/pkg/tir"
)

// TODO: see https://github.com/spf13/viper

func configEditor() edit.Editor {
	return edit.Tea{}
}

func configService() *tir.Service {
	return &tir.Service{Store: configStore()}
}

func configStore() store.Store {
	// Placeholder: include a record that can be updated or deleted.
	return store.NewMemory(&text.Text{
		ID:     "abc123de",
		URL:    "Initial URL",
		Title:  "Initial Title",
		Author: "Initial Author",
		Note:   "Initial Note",
	})
}
