package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPISecret(t *testing.T) {
	expectedSecret := "some-secret"
	t.Setenv("TIR_API_SECRET", expectedSecret)
	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, expectedSecret, cfg.GetAPISecret())
}
