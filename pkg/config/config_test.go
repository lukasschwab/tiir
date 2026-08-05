package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sethvargo/go-envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvironmentOverridesConfigFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ".tir.config")
	require.NoError(t, os.WriteFile(configPath, []byte(`{"store":{"type":"http","base_url":"https://example.test","api_secret":"from-file"}}`), 0o600))

	cfg, err := load(
		[]string{configPath},
		envconfig.MapLookuper(map[string]string{"TIR_API_SECRET": "from-env"}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, cfg.App.Close()) })
	assert.Equal(t, "from-env", cfg.GetAPISecret())
}

func TestLoadCommandLineOverridesEnvironment(t *testing.T) {
	secret := "from-flag"

	cfg, err := load(
		nil,
		envconfig.MapLookuper(map[string]string{
			"TIR_API_SECRET": "from-env",
			"TIR_TYPE":       "memory",
		}),
		Overrides{APISecret: &secret},
	)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, cfg.App.Close()) })
	assert.Equal(t, "memory", cfg.values.Store.Type)
	assert.Equal(t, "from-flag", cfg.GetAPISecret())
}
