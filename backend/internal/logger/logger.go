// Package logger builds the application's zap logger.
package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a zap logger with the given level ("debug", "info", "warn",
// "error") and format ("json" for containers, "console" for local dev).
func New(level, format string) (*zap.Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	var cfg zap.Config
	switch format {
	case "json":
		cfg = zap.NewProductionConfig()
	case "console":
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		return nil, fmt.Errorf("unsupported log format %q", format)
	}
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	// Request logs at Warn/Error would otherwise carry the middleware's own
	// stack, which says nothing useful. Panics log their stack explicitly.
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder // "1.2ms", not 0.0012

	log, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}
	return log, nil
}
