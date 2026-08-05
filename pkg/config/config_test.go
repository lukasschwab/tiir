package config

import (
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
				"TIR_TYPE":       "memory",
			}),
			envconfig.MapLookuper(map[string]string{"TIR_API_SECRET": "from-fallback"}),
		},
	)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, cfg.App.Close()) })
	assert.Equal(t, "memory", cfg.values.Store.Type)
	assert.Equal(t, "from-primary", cfg.GetAPISecret())
}
