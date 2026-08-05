package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/lukasschwab/tiir/pkg/config"
	"github.com/sethvargo/go-envconfig"
)

// CLI describes tir's command line interface.
type CLI struct {
	Verbose bool `short:"v" help:"Enable verbose logging."`

	Store            *string `short:"s" enum:"file,memory,http,libsql" help:"Store to use (file, memory, http, libsql)."`
	FileLocation     *string `name:"file-location" help:"File to use when store is file."`
	BaseURL          *string `name:"base-url" help:"Service URL to use when store is http."`
	APISecret        *string `name:"api-secret" help:"API secret to use when store is http."`
	ConnectionString *string `name:"connection-string" help:"Connection string to use when store is libsql."`
	Editor           *string `short:"e" enum:"vim,tea,huh" help:"Editor to use (vim, tea, huh)."`

	Create  CreateCommand  `cmd:"" help:"Record a text you read."`
	List    ListCommand    `cmd:"" help:"List all the texts you recorded reading."`
	Update  UpdateCommand  `cmd:"" aliases:"edit" help:"Update your record of a text you read."`
	Delete  DeleteCommand  `cmd:"" help:"Delete your record of a text you read."`
	Migrate MigrateCommand `cmd:"" help:"Batch-create records from an existing tir HTML file."`
}

type runtime struct {
	cfg    *config.Config
	stdout io.Writer
}

func (cli CLI) configLookuper() envconfig.Lookuper {
	values := make(map[string]string)
	put := func(key string, value *string) {
		if value != nil {
			values[key] = *value
		}
	}

	// Set both spellings so a CLI flag outranks either supported environment
	// variable alias.
	put("TIR_TYPE", cli.Store)
	put("TIR_STORE_TYPE", cli.Store)
	put("TIR_STORE_PATH", cli.FileLocation)
	put("TIR_STORE_BASE_URL", cli.BaseURL)
	put("TIR_API_SECRET", cli.APISecret)
	put("TIR_CONNECTION_STRING", cli.ConnectionString)
	put("TIR_EDITOR", cli.Editor)
	return envconfig.MapLookuper(values)
}

// Execute parses and executes the tir CLI.
func Execute() {
	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("tir"),
		kong.Description(`tir ('Today I Read...') logs the articles you read.

By default, it writes a JSON collection to $HOME/.tir.json with an interactive
CLI for adding readings. Configure tir with /etc/tir/.tir.config,
$HOME/.tir.config, TIR_* environment variables, or the flags below.`),
		kong.UsageOnError(),
	)
	if err != nil {
		log.Fatalf("build command line parser: %v", err)
	}

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	if !cli.Verbose {
		log.SetOutput(io.Discard)
	}

	cfg, err := config.Load(cli.configLookuper(), envconfig.OsLookuper())
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	defer func() {
		if err := cfg.App.Close(); err != nil {
			log.Printf("error closing app: %v", err)
		}
	}()

	if err := ctx.Run(&runtime{cfg: cfg, stdout: os.Stdout}); err != nil {
		log.Fatalf("%v", err)
	}
}

func optionList(options []string) string { return strings.Join(options, ", ") }
func invalidOption(name, value string, options []string) error {
	return fmt.Errorf("invalid %s %q; use one of %s", name, value, optionList(options))
}
