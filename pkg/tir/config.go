package tir

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/lukasschwab/tiir/pkg/edit"
	"github.com/lukasschwab/tiir/pkg/store"
	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/spf13/viper"
)

// Config keys.
const (
	// ConfigStore is the top-level key for store configuration.
	ConfigStore = "store"
	// ConfigStoreType must match a StoreType.
	ConfigStoreType = ConfigStore + ".type"

	// ConfigFileStoreLocation must be defined for file stores.
	ConfigFileStoreLocation = ConfigStore + ".path"

	// ConfigHTTPStoreBaseURL must be defined for HTTP stores.
	ConfigHTTPStoreBaseURL = ConfigStore + ".base_url"
	// ConfigHTTPStoreAPISecret defines an API secret to authorize requests to
	// the tir server at base_url. This is an optional config variable, but a
	// server that requires it will reject requests.
	ConfigHTTPStoreAPISecret = ConfigStore + ".api_secret"

	// ConfigEditor is the top-level key for CLI editor configuration.
	ConfigEditor = "editor"
)

type StoreType string

const (
	// StoreTypeFile selects the store.file store (default).
	StoreTypeFile StoreType = "file"
	// StoreTypeMemory selects the store.memory store.
	StoreTypeMemory StoreType = "memory"
	// StoreTypeMemory selects the store.http store.
	StoreTypeHTTP StoreType = "http"
)

// StoreOptions group the available StoreTypes for rendering CLI helper text.
var StoreOptions = []string{string(StoreTypeFile), string(StoreTypeMemory), string(StoreTypeHTTP)}

var storeFactories = map[StoreType]func() (store.Store, error){
	StoreTypeFile: func() (store.Store, error) {
		filepath := viper.GetString(ConfigFileStoreLocation)
		if filepath == "" {
			return nil, errors.New("must provide filepath for file store")
		}
		log.Printf("Using file store: %v", filepath)
		return store.UseFile(filepath)
	},
	StoreTypeMemory: func() (store.Store, error) {
		log.Printf("Using memory store")
		return store.UseMemory(), nil
	},
	StoreTypeHTTP: func() (store.Store, error) {
		baseURL := viper.GetString(ConfigHTTPStoreBaseURL)
		if baseURL == "" {
			return nil, errors.New("must provide base URL for HTTP store")
		}
		log.Printf("Using HTTP store: %v", baseURL)

		apiSecret := viper.GetString(ConfigHTTPStoreAPISecret)
		if apiSecret == "" {
			log.Printf("No API secret provided; store may reject requests")
		}
		return store.UseHTTP(baseURL, apiSecret)
	},
}

type EditorType string

const (
	// EditorTypeVim selects the edit.Vim editor.
	EditorTypeVim EditorType = "vim"
	// EditorTypeTea selects the edit.Tea editor (default).
	EditorTypeTea EditorType = "tea"
)

var (
	// EditorOptions render the list of options in CLI help text.
	EditorOptions = []string{string(EditorTypeVim), string(EditorTypeTea)}
	editors       = map[EditorType]text.Editor{
		EditorTypeVim: edit.Vim,
		EditorTypeTea: edit.Tea,
	}
)

// FromConfig loads a tir configuration from user-provided configuration.
// Users can provide configuration via a JSON config file, via environment
// variables, or through command-line arguments with the appropriate viper
// bindings.
//
// + The default Service is backed by a file at $HOME/.tir.json.
// + The default text.Editor is edit.Tea.
//
// The caller is responsible for calling (Config).Service.Close() appropriately.
func LoadConfig() (Config, error) {
	viper.SetEnvPrefix("tir")
	viper.SetConfigName(".tir.config")
	viper.SetConfigType("json")
	viper.AddConfigPath("/etc/tir/")

	if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(home)
		viper.SetDefault(ConfigFileStoreLocation, home+"/.tir.json")
	}

	// Write enum-type results as strings to avoid silently borking
	// viper.GetString's type indirection.
	viper.SetDefault(ConfigStoreType, string(StoreTypeFile))
	viper.SetDefault(ConfigEditor, string(EditorTypeTea))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Printf("no config file")
		} else {
			return Config{}, fmt.Errorf("can't read config file: %w", err)
		}
	}

	cfg := Config{v: viper.GetViper()}

	// Construct a service.
	storeFactory, ok := storeFactories[cfg.getStoreType()]
	if !ok {
		return cfg, fmt.Errorf("invalid store type '%v'", cfg.getStoreType())
	}
	store, err := storeFactory()
	if err != nil {
		return cfg, fmt.Errorf("error generating store: %w", err)
	}
	cfg.Service = New(store)

	// Construct a store.
	if cfg.Editor, ok = editors[cfg.getEditorType()]; !ok {
		return cfg, fmt.Errorf("invalid editor type '%v'", cfg.getEditorType())
	}

	return cfg, nil
}

// NOTE: should Config embed a Service and an Editor, or should it just store
// the factory for the Service? Embedding is nice for getting a single package-
// scoped variable for the cobra app (magic I want to minimize), but I'd prefer
// to isolate store-factory errors from config-loading errors.
type Config struct {
	v       *viper.Viper
	Service *Service
	Editor  text.Editor
}

func (cfg *Config) getStoreType() StoreType {
	return StoreType(cfg.v.GetString(ConfigStoreType))
}

func (cfg *Config) getEditorType() EditorType {
	return EditorType(cfg.v.GetString(ConfigEditor))
}
