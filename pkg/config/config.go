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
	lookuper := envconfig.MultiLookuper(lookupers...)

	v := defaultValues()
	for _, path := range paths {
		if err := applyFile(&v, path); err != nil {
			return nil, err
		}
	}

	var env envValues
	if err := envconfig.ProcessWith(context.Background(), &envconfig.Config{
		Target:   &env,
		Lookuper: lookuper,
	}); err != nil {
		return nil, fmt.Errorf("read environment: %w", err)
	}
	applyEnv(&v, env)

	cfg := &Config{values: v}
	if err := cfg.initialize(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func defaultValues() values {
	v := values{Editor: string(EditorTypeTea)}
	v.Store.Type = string(StoreTypeFile)
	if home, err := os.UserHomeDir(); err == nil {
		v.Store.Path = filepath.Join(home, ".tir.json")
	}
	return v
}

func configPaths() []string {
	paths := []string{"/etc/tir/.tir.config"}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".tir.config"))
	}
	return paths
}

func applyFile(v *values, path string) error {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read config file %q: %w", path, err)
	}
	var file fileValues
	if err := json.Unmarshal(contents, &file); err != nil {
		return fmt.Errorf("parse config file %q: %w", path, err)
	}
	if file.Store != nil {
		apply(&v.Store.Type, file.Store.Type)
		apply(&v.Store.Path, file.Store.Path)
		apply(&v.Store.BaseURL, file.Store.BaseURL)
		apply(&v.Store.APISecret, file.Store.APISecret)
		apply(&v.Store.ConnectionString, file.Store.ConnectionString)
	}
	apply(&v.Editor, file.Editor)
	return nil
}

func applyEnv(v *values, env envValues) {
	// TIR_TYPE is the legacy Viper spelling. TIR_STORE_TYPE is accepted as a
	// more descriptive alias, with the latter taking precedence when both exist.
	apply(&v.Store.Type, env.StoreType)
	apply(&v.Store.Type, env.StoreTypeNested)
	apply(&v.Store.Path, env.FileLocation)
	apply(&v.Store.BaseURL, env.BaseURL)
	apply(&v.Store.APISecret, env.APISecret)
	apply(&v.Store.ConnectionString, env.ConnectionString)
	apply(&v.Editor, env.Editor)
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
