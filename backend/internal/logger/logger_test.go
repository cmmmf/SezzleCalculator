package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestNew(t *testing.T) {
	cases := []struct {
		level, format string
		enabled       zapcore.Level
		disabled      zapcore.Level
	}{
		{level: "info", format: "json", enabled: zapcore.InfoLevel, disabled: zapcore.DebugLevel},
		{level: "debug", format: "console", enabled: zapcore.DebugLevel, disabled: zapcore.DebugLevel - 1},
		{level: "error", format: "json", enabled: zapcore.ErrorLevel, disabled: zapcore.WarnLevel},
	}
	for _, tc := range cases {
		t.Run(tc.level+"/"+tc.format, func(t *testing.T) {
			log, err := New(tc.level, tc.format)
			require.NoError(t, err)
			assert.True(t, log.Core().Enabled(tc.enabled))
			assert.False(t, log.Core().Enabled(tc.disabled))
		})
	}
}

func TestNewInvalid(t *testing.T) {
	_, err := New("loud", "json")
	assert.Error(t, err)

	_, err = New("info", "xml")
	assert.Error(t, err)
}
