package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func envFrom(m map[string]string) func(string) string {
	return func(key string) string { return m[key] }
}

func TestLoadFromDefaults(t *testing.T) {
	cfg, err := LoadFrom(envFrom(nil))
	require.NoError(t, err)
	assert.Equal(t, Config{
		Port:        DefaultPort,
		GinMode:     DefaultGinMode,
		LogLevel:    DefaultLogLevel,
		LogFormat:   DefaultLogFormat,
		CORSOrigins: []string{DefaultCORSOrigins},
	}, cfg)
	assert.Equal(t, ":8080", cfg.Addr())
}

func TestLoadFromOverrides(t *testing.T) {
	cfg, err := LoadFrom(envFrom(map[string]string{
		"PORT":         " 9090 ",
		"GIN_MODE":     "debug",
		"LOG_LEVEL":    "DEBUG",
		"LOG_FORMAT":   "Console",
		"CORS_ORIGINS": "http://a.test, ,http://b.test",
	}))
	require.NoError(t, err)
	assert.Equal(t, Config{
		Port:        "9090",
		GinMode:     "debug",
		LogLevel:    "debug",
		LogFormat:   "console",
		CORSOrigins: []string{"http://a.test", "http://b.test"},
	}, cfg)
}

func TestLoadFromInvalid(t *testing.T) {
	cases := map[string]map[string]string{
		"non-numeric port":  {"PORT": "abc"},
		"port out of range": {"PORT": "70000"},
		"port zero":         {"PORT": "0"},
		"bad gin mode":      {"GIN_MODE": "production"},
		"bad log level":     {"LOG_LEVEL": "verbose"},
		"bad log format":    {"LOG_FORMAT": "xml"},
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := LoadFrom(envFrom(env))
			assert.Error(t, err)
		})
	}
}

func TestLoadReadsProcessEnvironment(t *testing.T) {
	t.Setenv("PORT", "7070")
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "7070", cfg.Port)
}
