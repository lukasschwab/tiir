// Package config loads a tir configuration, either from a JSON file
// (e.g. $HOME/.tir.config), from command-line arguments if the config variables
// are bound with [viper.BindPFlag], or from environment variables (though the
// latter are not explicitly supported).
//
// Loading a [Config] generates a [tir.Interface] for interacting with the
// configured store and a [text.Editor] for getting user input.
//
// # Defaults
//
// In the absence of explicit configuration, [Load] yields a [Config] backed by
// a [store.File] store pointing at $HOME/.tir.json and an [edit.Tea] editor.
package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lukasschwab/tiir/pkg/edit"
	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/lukasschwab/tiir/pkg/tir"
	"github.com/spf13/viper"
)

// Values providable via a JSON config file. For example, this .tir.config file
// configures tir to use a file store rooted at /Users/me/.tir.json and the Tea
// editor:
//
//	{ "store": { "type": "file", "path": "/Users/me/.tir.json" }, "editor": "tea" }
//
// For comparison, this .tir.config file configures tir to talk to a server at
// tir.example.com:
//
//	{ "store": { "type": "http", "base_url": "https://tir.example.com", "api_secret": "YOUR_SECRET" } }
//
// And this .tir.config file configures tir to use a Turso-hosted database:
//
//	{ "store": { "type": "libsql", "connection_string": "libsql://[your-database].turso.io?authToken=[your-auth-token]" } }
//
// For info on where tir looks for a config, see [Load]. For info about
// how to provide configuration, see [viper].
//
// [viper]: https://github.com/spf13/viper
const (
	// KeyStoreGroup is the top-level key for store configuration.
	KeyStoreGroup = "store"
	// KeyStoreType must match a StoreType.
	KeyStoreType = KeyStoreGroup + ".type"

	// KeyFileStoreLocation must be defined for file stores.
	KeyFileStoreLocation = KeyStoreGroup + ".path"

	// KeyHTTPStoreBaseURL must be defined for HTTP stores.
	KeyHTTPStoreBaseURL = KeyStoreGroup + ".base_url"
	// KeyHTTPStoreAPISecret defines an API secret to authorize requests to
	// the tir server at base_url. This is an optional config variable, but a
	// server that requires it will reject requests.
	KeyHTTPStoreAPISecret = KeyStoreGroup + ".api_secret"

	// KeyLibSQLStoreConnectionString must be defined for LibSQL stores. It
	// specifies where/how to connect to LibSQL. For example:
	//
	// + A local SQLite file:	"file://path/to/file.db"
	// + A local sqld instance:	"http://127.0.0.1:8080"
	// + A Turso-hosted DB:		"libsql://[your-database].turso.io?authToken=[your-auth-token]"
	KeyLibSQLStoreConnectionString = KeyStoreGroup + ".connection_string"

	// KeyEditor is the top-level key for CLI editor configuration.
	KeyEditor = "editor"
)

type storeType string

// Possible values for the store.type config variable. Providing any other store
// type value will cause [Load] to fail.
const (
	// StoreTypeFile selects the store.file store (default).
	StoreTypeFile storeType = "file"
	// StoreTypeMemory selects the store.memory store.
	StoreTypeMemory storeType = "memory"
	// StoreTypeMemory selects the store.http store.
	StoreTypeHTTP storeType = "http"
	// StoreTypeLibSQL selectes the store.libsql store (supports Turso).
	StoreTypeLibSQL storeType = "libsql"
)

type editorType string

// Possible values for the editor config variable. Providing any other editor
// value will cause [Load] to fail.
const (
	// EditorTypeVim selects the edit.Vim editor.
	EditorTypeVim editorType = "vim"
	// EditorTypeTea selects the edit.Tea editor (default).
	EditorTypeTea editorType = "tea"
)

// Enum-option to value lookups.
var (
	storeFactories = map[storeType]func(*Config) (store.Interface, error){
		StoreTypeFile: func(*Config) (store.Interface, error) {
			filepath := viper.GetString(KeyFileStoreLocation)
			if filepath == "" {
				return nil, errors.New("must provide filepath for file store")
			}
			log.Printf("Using file store: %v", filepath)
			return store.UseFile(filepath)
		},
		StoreTypeMemory: func(*Config) (store.Interface, error) {
			log.Printf("Using memory store")
			return store.UseMemory(), nil
		},
		StoreTypeHTTP: func(cfg *Config) (store.Interface, error) {
			baseURL := viper.GetString(KeyHTTPStoreBaseURL)
			if baseURL == "" {
				return nil, errors.New("must provide base URL for HTTP store")
			}
			log.Printf("Using HTTP store: %v", baseURL)

			apiSecret := cfg.GetAPISecret()
			if apiSecret == "" {
				log.Printf("No API secret provided; store may reject requests")
			}
			return store.UseHTTP(baseURL, apiSecret)
		},
		StoreTypeLibSQL: func(cfg *Config) (store.Interface, error) {
			connectionString := viper.GetString(KeyLibSQLStoreConnectionString)
			if connectionString == "" {
				return nil, errors.New("must provide connection string for LibSQL store")
			}
			log.Printf("Using LibSQL store")

			return store.UseLibSQL(connectionString)
		},
	}

	editors = map[editorType]text.Editor{
		EditorTypeVim: edit.Vim,
		EditorTypeTea: edit.Tea,
	}
)

// Load loads a tir configuration from user-provided configuration.
// Users can provide configuration via a JSON config file, via environment
// variables, or through command-line arguments with the appropriate viper
// bindings.
//
// The default [tir.Interface] is backed by a file at $HOME/.tir.json. The
// default [text.Editor] is [edit.Tea].
//
// The caller is responsible for calling [tir.Interface.Close] appropriately.
func Load() (*Config, error) {
	viper.SetEnvPrefix("tir")
	viper.SetConfigName(".tir.config")
	viper.SetConfigType("json")
	viper.AddConfigPath("/etc/tir/")

	if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(home)
		viper.SetDefault(KeyFileStoreLocation, home+"/.tir.json")
	}

	// Write enum-type results as strings to avoid silently borking
	// viper.GetString's type indirection.
	viper.SetDefault(KeyStoreType, string(StoreTypeFile))
	viper.SetDefault(KeyEditor, string(EditorTypeTea))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Printf("no config file")
		} else {
			return nil, fmt.Errorf("can't read config file: %w", err)
		}
	}

	// The standard post-prefixing key for this value is unwieldy
	// (e.g. TIR_STORE.API_SECRET), hence the replacement (e.g. TIR_API_SECRET).
	viper.SetEnvKeyReplacer(strings.NewReplacer("STORE.", ""))
	if err := viper.BindEnv(KeyStoreType); err != nil {
		log.Printf("ignoring error binding %v: %v", KeyStoreType, err)
	} else if err := viper.BindEnv(KeyHTTPStoreAPISecret); err != nil {
		log.Printf("ignoring error binding %v: %v", KeyHTTPStoreAPISecret, err)
	} else if err := viper.BindEnv(KeyLibSQLStoreConnectionString); err != nil {
		log.Printf("ignoring error binding %v: %v", KeyLibSQLStoreConnectionString, err)
	}

	cfg := &Config{v: viper.GetViper()}

	// Construct a service.
	storeFactory, ok := storeFactories[cfg.getStoreType()]
	if !ok {
		return cfg, fmt.Errorf("invalid store type '%v'", cfg.getStoreType())
	}
	store, err := storeFactory(cfg)
	if err != nil {
		return cfg, fmt.Errorf("error generating store: %w", err)
	}
	cfg.App = tir.New(store)

	// Construct a store.
	if cfg.Editor, ok = editors[cfg.getEditorType()]; !ok {
		return cfg, fmt.Errorf("invalid editor type '%v'", cfg.getEditorType())
	}

	return cfg, nil
}

// Config for a Service and Editor. Use [Load] to construct a Config that
// respects the [viper]-provided user configuration environment.
type Config struct {
	v      *viper.Viper
	App    tir.Interface
	Editor text.Editor
}

func (cfg *Config) getStoreType() storeType {
	return storeType(cfg.v.GetString(KeyStoreType))
}

func (cfg *Config) getEditorType() editorType {
	return editorType(cfg.v.GetString(KeyEditor))
}

// GetAPISecret provided through [viper] or the env variable TIR_API_SECRET if
// it's present.
func (cfg *Config) GetAPISecret() string {
	return cfg.v.GetString(KeyHTTPStoreAPISecret)
}
