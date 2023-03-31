package cmd

import (
	"github.com/lukasschwab/tiir/pkg/edit"
	"github.com/lukasschwab/tiir/pkg/tir"
)

// Config properties initialized and closed by rootCmd pre- and post-run funcs.
var (
	configuredService *tir.Service
	configuredEditor  edit.Editor
)
