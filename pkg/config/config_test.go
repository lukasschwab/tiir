package config

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/sethvargo/go-envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadLookupersOverrideConfigFileInPriorityOrder(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ".tir.config")
	require.NoError(t, os.WriteFile(configPath, []byte(`{"store":{"type":"http","base_url":"https://example.test","api_secret":"from-file"}}`), 0o600))

	cfg, err := load(
		[]string{configPath},
		[]envconfig.Lookuper{
			envconfig.MapLookuper(map[string]string{"TIR_API_SECRET": "from-primary"}),
			envconfig.MapLookuper(map[string]string{"TIR_API_SECRET": "from-fallback"}),
		},
	)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, cfg.App.Close()) })
	assert.Equal(t, "from-primary", cfg.GetAPISecret())
}

func TestMaskedJSONMasksAPISecret(t *testing.T) {
	cfg, err := load(
		nil,
		[]envconfig.Lookuper{envconfig.MapLookuper(map[string]string{
			"TIR_STORE_TYPE":        "http",
			"TIR_STORE_BASE_URL":    "https://example.test",
			"TIR_API_SECRET":        "not-for-output",
			"TIR_CONNECTION_STRING": "libsql://example.turso.io?authToken=also-not-for-output",
		})},
	)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, cfg.App.Close()) })

	contents, err := cfg.MaskedJSON()
	require.NoError(t, err)

	var output values
	require.NoError(t, json.Unmarshal(contents, &output))
	assert.Equal(t, "http", output.Store.Type)
	assert.Equal(t, cfg.values.Store.Path, output.Store.Path)
	assert.Equal(t, "https://example.test", output.Store.BaseURL)
	assert.Equal(t, maskedSecret, output.Store.APISecret)
	connectionString, err := url.Parse(output.Store.ConnectionString)
	require.NoError(t, err)
	assert.Equal(t, maskedSecret, connectionString.Query().Get("authToken"))
	assert.Equal(t, "tea", output.Editor)
	assert.NotContains(t, string(contents), "not-for-output")
	assert.NotContains(t, string(contents), "also-not-for-output")
}

func TestLoadLaterConfigFileOverridesEarlierConfigFile(t *testing.T) {
	configDir := t.TempDir()
	firstPath := filepath.Join(configDir, "first.json")
	secondPath := filepath.Join(configDir, "second.json")
	require.NoError(t, os.WriteFile(firstPath, []byte(`{"store":{"type":"http","base_url":"https://example.test","api_secret":"from-first"}}`), 0o600))
	require.NoError(t, os.WriteFile(secondPath, []byte(`{"store":{"api_secret":"from-second"}}`), 0o600))

	cfg, err := load([]string{firstPath, secondPath}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, cfg.App.Close()) })
	assert.Equal(t, "from-second", cfg.GetAPISecret())
}

func TestLoadHigherPriorityLookuperWins(t *testing.T) {
	cfg, err := load(
		nil,
		[]envconfig.Lookuper{
			envconfig.MapLookuper(map[string]string{
				"TIR_API_SECRET": "from-primary",
				"TIR_STORE_TYPE": "memory",
			}),
			envconfig.MapLookuper(map[string]string{"TIR_API_SECRET": "from-fallback"}),
		},
	)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, cfg.App.Close()) })
	assert.Equal(t, "memory", cfg.values.Store.Type)
	assert.Equal(t, "from-primary", cfg.GetAPISecret())
}
