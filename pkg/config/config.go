// Package config loads tir configuration from JSON files, environment variables,
// and command-line overrides.
package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/lukasschwab/tiir/pkg/edit"
	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/lukasschwab/tiir/pkg/tir"
	"github.com/sethvargo/go-envconfig"
)

const (
	KeyStoreGroup                  = "store"
	KeyStoreType                   = KeyStoreGroup + ".type"
	KeyFileStoreLocation           = KeyStoreGroup + ".path"
	KeyHTTPStoreBaseURL            = KeyStoreGroup + ".base_url"
	KeyHTTPStoreAPISecret          = KeyStoreGroup + ".api_secret"
	KeyLibSQLStoreConnectionString = KeyStoreGroup + ".connection_string"
	KeyEditor                      = "editor"
)

type storeType string

const (
	StoreTypeFile   storeType = "file"
	StoreTypeMemory storeType = "memory"
	StoreTypeHTTP   storeType = "http"
	StoreTypeLibSQL storeType = "libsql"
)

type editorType string

const (
	EditorTypeVim editorType = "vim"
	EditorTypeTea editorType = "tea"
	EditorTypeHuh editorType = "huh"
)

type values struct {
	Store struct {
		Type             string `json:"type"`
		Path             string `json:"path"`
		BaseURL          string `json:"base_url"`
		APISecret        string `json:"api_secret"`
		ConnectionString string `json:"connection_string"`
	} `json:"store"`
	Editor string `json:"editor"`
}

type fileValues struct {
	Store *struct {
		Type             *string `json:"type"`
		Path             *string `json:"path"`
		BaseURL          *string `json:"base_url"`
		APISecret        *string `json:"api_secret"`
		ConnectionString *string `json:"connection_string"`
	} `json:"store"`
	Editor *string `json:"editor"`
}

type fileLookuper map[string]string

func (lookuper fileLookuper) Lookup(key string) (string, bool) {
	value, ok := lookuper[key]
	return value, ok
}

func newFileLookuper(path string) (envconfig.Lookuper, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}

	var file fileValues
	if err := json.Unmarshal(contents, &file); err != nil {
		return nil, fmt.Errorf("parse config file %q: %w", path, err)
	}

	values := make(fileLookuper)
	put := func(key string, value *string) {
		if value != nil {
			values[key] = *value
		}
	}
	if file.Store != nil {
		put("TIR_TYPE", file.Store.Type)
		put("TIR_STORE_PATH", file.Store.Path)
		put("TIR_STORE_BASE_URL", file.Store.BaseURL)
		put("TIR_API_SECRET", file.Store.APISecret)
		put("TIR_CONNECTION_STRING", file.Store.ConnectionString)
	}
	put("TIR_EDITOR", file.Editor)
	return values, nil
}

type envValues struct {
	StoreType        *string `env:"TIR_TYPE,noinit"`
	StoreTypeNested  *string `env:"TIR_STORE_TYPE,noinit"`
	FileLocation     *string `env:"TIR_STORE_PATH,noinit"`
	BaseURL          *string `env:"TIR_STORE_BASE_URL,noinit"`
	APISecret        *string `env:"TIR_API_SECRET,noinit"`
	ConnectionString *string `env:"TIR_CONNECTION_STRING,noinit"`
	Editor           *string `env:"TIR_EDITOR,noinit"`
}

// Load constructs a configured application from the supplied lookupers. Values
// are applied in this order: defaults, /etc/tir/.tir.config,
// $HOME/.tir.config, and lookupers in priority order.
func Load(lookuper envconfig.Lookuper, fallbacks ...envconfig.Lookuper) (*Config, error) {
	return load(configPaths(), append([]envconfig.Lookuper{lookuper}, fallbacks...))
}

// load constructs a configured application using the supplied configuration
// paths and environment lookupers. Lookupers are evaluated in priority order,
// letting tests use isolated files and map-backed values rather than mutating
// process state.
func load(paths []string, lookupers []envconfig.Lookuper) (*Config, error) {
	sources := append([]envconfig.Lookuper{}, lookupers...)
	for i := len(paths) - 1; i >= 0; i-- {
		lookuper, err := newFileLookuper(paths[i])
		if err != nil {
			return nil, err
		}
		if lookuper != nil {
			sources = append(sources, lookuper)
		}
	}
	sources = append(sources, defaultLookuper())

	var env envValues
	if err := envconfig.ProcessWith(context.Background(), &envconfig.Config{
		Target:   &env,
		Lookuper: envconfig.MultiLookuper(sources...),
	}); err != nil {
		return nil, fmt.Errorf("read environment: %w", err)
	}

	cfg := &Config{values: valuesFrom(env)}
	if err := cfg.initialize(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func defaultLookuper() envconfig.Lookuper {
	values := map[string]string{
		"TIR_TYPE":   string(StoreTypeFile),
		"TIR_EDITOR": string(EditorTypeTea),
	}
	if home, err := os.UserHomeDir(); err == nil {
		values["TIR_STORE_PATH"] = filepath.Join(home, ".tir.json")
	}
	return envconfig.MapLookuper(values)
}

func configPaths() []string {
	paths := []string{"/etc/tir/.tir.config"}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".tir.config"))
	}
	return paths
}

func valuesFrom(env envValues) values {
	var values values

	// TIR_TYPE is the legacy Viper spelling. TIR_STORE_TYPE is accepted as a
	// more descriptive alias, with the latter taking precedence when both exist.
	apply(&values.Store.Type, env.StoreType)
	apply(&values.Store.Type, env.StoreTypeNested)
	apply(&values.Store.Path, env.FileLocation)
	apply(&values.Store.BaseURL, env.BaseURL)
	apply(&values.Store.APISecret, env.APISecret)
	apply(&values.Store.ConnectionString, env.ConnectionString)
	apply(&values.Editor, env.Editor)
	return values
}

func apply(destination *string, source *string) {
	if source != nil {
		*destination = *source
	}
}

// Config contains the configured store-backed app and editor. Callers must
// close App when finished.
type Config struct {
	values values
	App    tir.Interface
	Editor text.Editor
}

func (cfg *Config) initialize() error {
	switch storeType(cfg.values.Store.Type) {
	case StoreTypeFile:
		if cfg.values.Store.Path == "" {
			return errors.New("must provide filepath for file store")
		}
		log.Printf("Using file store: %v", cfg.values.Store.Path)
		appStore, err := store.UseFile(cfg.values.Store.Path)
		if err != nil {
			return fmt.Errorf("create file store: %w", err)
		}
		cfg.App = tir.New(appStore)
	case StoreTypeMemory:
		log.Printf("Using memory store")
		cfg.App = tir.New(store.UseMemory())
	case StoreTypeHTTP:
		if cfg.values.Store.BaseURL == "" {
			return errors.New("must provide base URL for HTTP store")
		}
		log.Printf("Using HTTP store: %v", cfg.values.Store.BaseURL)
		appStore, err := store.UseHTTP(cfg.values.Store.BaseURL, cfg.GetAPISecret())
		if err != nil {
			return fmt.Errorf("create HTTP store: %w", err)
		}
		cfg.App = tir.New(appStore)
	case StoreTypeLibSQL:
		if cfg.values.Store.ConnectionString == "" {
			return errors.New("must provide connection string for LibSQL store")
		}
		log.Printf("Using LibSQL store")
		appStore, err := store.UseLibSQL(cfg.values.Store.ConnectionString)
		if err != nil {
			return fmt.Errorf("create LibSQL store: %w", err)
		}
		cfg.App = tir.New(appStore)
	default:
		return fmt.Errorf("invalid store type %q", cfg.values.Store.Type)
	}

	switch editorType(cfg.values.Editor) {
	case EditorTypeVim:
		cfg.Editor = edit.Vim
	case EditorTypeTea:
		cfg.Editor = edit.Tea
	case EditorTypeHuh:
		cfg.Editor = edit.Huh
	default:
		return fmt.Errorf("invalid editor type %q", cfg.values.Editor)
	}
	return nil
}

// GetAPISecret returns the configured API secret.
func (cfg *Config) GetAPISecret() string { return cfg.values.Store.APISecret }
