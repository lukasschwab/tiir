package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvironmentOverridesConfigFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".tir.config"), []byte(`{"store":{"type":"http","base_url":"https://example.test","api_secret":"from-file"}}`), 0o600))
	t.Setenv("TIR_API_SECRET", "from-env")

	cfg, err := Load()
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, cfg.App.Close()) })
	assert.Equal(t, "from-env", cfg.GetAPISecret())
}

func TestLoadCommandLineOverridesEnvironment(t *testing.T) {
	t.Setenv("TIR_API_SECRET", "from-env")
	t.Setenv("TIR_TYPE", "memory")
	secret := "from-flag"

	cfg, err := Load(Overrides{APISecret: &secret})
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, cfg.App.Close()) })
	assert.Equal(t, "memory", cfg.values.Store.Type)
	assert.Equal(t, "from-flag", cfg.GetAPISecret())
}
