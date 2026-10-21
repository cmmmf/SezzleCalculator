// Package config loads the server configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds the runtime configuration of the server.
type Config struct {
	Port        string   // PORT: TCP port to listen on.
	GinMode     string   // GIN_MODE: debug | release | test.
	LogLevel    string   // LOG_LEVEL: debug | info | warn | error.
	LogFormat   string   // LOG_FORMAT: json | console.
	CORSOrigins []string // CORS_ORIGINS: comma-separated list of allowed origins.
}

// Default values used when a variable is unset or empty.
const (
	DefaultPort        = "8080"
	DefaultGinMode     = "release"
	DefaultLogLevel    = "info"
	DefaultLogFormat   = "json"
	DefaultCORSOrigins = "http://localhost:5173"
)

// Load reads the configuration from the process environment.
func Load() (Config, error) {
	return LoadFrom(os.Getenv)
}

// LoadFrom reads the configuration using getenv, which makes it testable
// without touching the real environment.
func LoadFrom(getenv func(string) string) (Config, error) {
	cfg := Config{
		Port:        valueOr(getenv("PORT"), DefaultPort),
		GinMode:     valueOr(getenv("GIN_MODE"), DefaultGinMode),
		LogLevel:    strings.ToLower(valueOr(getenv("LOG_LEVEL"), DefaultLogLevel)),
		LogFormat:   strings.ToLower(valueOr(getenv("LOG_FORMAT"), DefaultLogFormat)),
		CORSOrigins: splitList(valueOr(getenv("CORS_ORIGINS"), DefaultCORSOrigins)),
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid PORT %q: must be a number between 1 and 65535", c.Port)
	}
	switch c.GinMode {
	case "debug", "release", "test":
	default:
		return fmt.Errorf("invalid GIN_MODE %q: must be debug, release or test", c.GinMode)
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid LOG_LEVEL %q: must be debug, info, warn or error", c.LogLevel)
	}
	switch c.LogFormat {
	case "json", "console":
	default:
		return fmt.Errorf("invalid LOG_FORMAT %q: must be json or console", c.LogFormat)
	}
	return nil
}

// Addr returns the listen address, e.g. ":8080".
func (c Config) Addr() string {
	return ":" + c.Port
}

func valueOr(v, fallback string) string {
	if v = strings.TrimSpace(v); v == "" {
		return fallback
	}
	return v
}

func splitList(s string) []string {
	var out []string
	for part := range strings.SplitSeq(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
