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

// Scratchpad config example.
/* {
	// Stores.
	"store": {
		"type": "file",
		"path": "~/.tir.json"
	},
	"store": {
		"type": "http",
		"base_url": "tir.fly.io",
	},
	"store": {
		"type": "memory",
	},

	// Editors.
	"editor": "vim",
	"editor": "tea",
} */

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

var StoreOptions = []string{string(StoreTypeFile), string(StoreTypeMemory), string(StoreTypeHTTP)}

var StoreFactories = map[StoreType]func() (store.Store, error){
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

var EditorOptions = []string{string(EditorTypeVim), string(EditorTypeTea)}

var Editors = map[EditorType]text.Editor{
	EditorTypeVim: edit.Vim,
	EditorTypeTea: edit.Tea,
}

// FromConfig loads a Service and text.Editor from defaults, overridden by user-
// provided configuration.
//
// + The default Service is backed by a file at $HOME/.tir.json.
// + The default text.Editor is edit.Tea.
func FromConfig() (*Service, text.Editor, error) {
	viper.SetConfigName(".tir.config")
	viper.SetConfigType("json")
	viper.AddConfigPath("/etc/tir/")

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	viper.AddConfigPath(home)

	// Write enum-type results as strings to avoid silently borking
	// viper.GetString's type indirection.
	viper.SetDefault(ConfigStoreType, string(StoreTypeFile))
	viper.SetDefault(ConfigFileStoreLocation, home+"/.tir.json")

	viper.SetDefault(ConfigEditor, string(EditorTypeTea))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Printf("no config file")
		} else {
			log.Fatalf("can't read config file: %v", err)
		}
	}

	storeType := viper.GetString(ConfigStoreType)

	var s store.Store
	if storeFactory, ok := StoreFactories[StoreType(storeType)]; !ok {
		return nil, nil, fmt.Errorf("invalid store type '%v'", storeType)
	} else if s, err = storeFactory(); err != nil {
		return nil, nil, fmt.Errorf("error generating %v store: %w", storeType, err)
	}

	editor := viper.GetString(ConfigEditor)
	if _, ok := Editors[EditorType(editor)]; !ok {
		return nil, nil, fmt.Errorf("invalid editor type '%v'", editor)
	}

	return New(s), Editors[EditorType(editor)], nil
}
