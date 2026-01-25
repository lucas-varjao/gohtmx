// backend/internal/config/config_test.go

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestConfigDir creates a temp dir with app.yml and returns the dir and a cleanup function.
// Tests use LoadConfigFromPath(dir) so ./configs is never touched.
func setupTestConfigDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "gohtmx-config-test-*")
	require.NoError(t, err)

	configContent := `
server:
  port: 8080
database:
  dsn: "test.db"
jwt:
  secret-key: "test-secret-key"
  access_token_ttl: 15m
  refresh_token_ttl: 24h
  issuer: "gohtmx-test"
`
	err = os.WriteFile(filepath.Join(dir, "app.yml"), []byte(configContent), 0644)
	require.NoError(t, err)

	cleanup := func() {
		_ = os.RemoveAll(dir)
		viper.Reset()
		cfg = nil
	}
	return dir, cleanup
}

func TestLoadConfig(t *testing.T) {
	dir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	c, err := LoadConfigFromPath(dir)
	require.NoError(t, err)
	require.NotNil(t, c)

	assert.Equal(t, 8080, c.Server.Port)
	assert.Equal(t, "test.db", c.Database.DSN)
	assert.Equal(t, "test-secret-key", c.JWT.SecretKey)
	assert.Equal(t, 15*time.Minute, c.JWT.AccessTokenTTL)
	assert.Equal(t, 24*time.Hour, c.JWT.RefreshTokenTTL)
	assert.Equal(t, "gohtmx-test", c.JWT.Issuer)
}

func TestLoadConfigError(t *testing.T) {
	viper.Reset()
	cfg = nil
	defer func() { viper.Reset(); cfg = nil }()

	// Dir exists but has no app.yml
	dir, err := os.MkdirTemp("", "gohtmx-config-empty-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	c, err := LoadConfigFromPath(dir)
	assert.Error(t, err)
	assert.Nil(t, c)
}

func TestGetConfig(t *testing.T) {
	dir, cleanup := setupTestConfigDir(t)
	defer cleanup()

	_, err := LoadConfigFromPath(dir)
	require.NoError(t, err)

	c := GetConfig()
	require.NotNil(t, c)
	assert.Equal(t, 8080, c.Server.Port)
}

func TestGetConfigBeforeLoad(t *testing.T) {
	viper.Reset()
	cfg = nil
	defer func() { viper.Reset(); cfg = nil }()

	assert.Nil(t, GetConfig())
}
