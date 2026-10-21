package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func init() { gin.SetMode(gin.TestMode) }

func serve(engine *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

// --- RequestID ---------------------------------------------------------------

func TestRequestIDGeneratedWhenMissing(t *testing.T) {
	engine := gin.New()
	engine.Use(RequestID())
	var seen string
	engine.GET("/", func(ctx *gin.Context) { seen = ctx.GetString(RequestIDKey) })

	rec := serve(engine, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Regexp(t, `^[0-9a-f]{32}$`, seen)
	assert.Equal(t, seen, rec.Header().Get(RequestIDHeader))
}

func TestRequestIDPropagatedWhenValid(t *testing.T) {
	engine := gin.New()
	engine.Use(RequestID())
	engine.GET("/", func(*gin.Context) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "abc-123_x.y")
	rec := serve(engine, req)

	assert.Equal(t, "abc-123_x.y", rec.Header().Get(RequestIDHeader))
}

func TestRequestIDReplacedWhenUnsafe(t *testing.T) {
	for name, id := range map[string]string{
		"log injection": "abc\nlevel=error",
		"too long":      strings.Repeat("a", 65),
		"spaces":        "a b",
	} {
		t.Run(name, func(t *testing.T) {
			engine := gin.New()
			engine.Use(RequestID())
			engine.GET("/", func(*gin.Context) {})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(RequestIDHeader, id)
			rec := serve(engine, req)

			assert.Regexp(t, `^[0-9a-f]{32}$`, rec.Header().Get(RequestIDHeader))
		})
	}
}

// --- Logger ------------------------------------------------------------------

func TestLoggerFieldsAndLevels(t *testing.T) {
	cases := []struct {
		status int
		level  zapcore.Level
	}{
		{http.StatusOK, zapcore.InfoLevel},
		{http.StatusNoContent, zapcore.InfoLevel},
		{http.StatusBadRequest, zapcore.WarnLevel},
		{http.StatusUnprocessableEntity, zapcore.WarnLevel},
		{http.StatusInternalServerError, zapcore.ErrorLevel},
	}
	for _, tc := range cases {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			core, logs := observer.New(zapcore.DebugLevel)
			engine := gin.New()
			engine.Use(RequestID(), Logger(zap.New(core)))
			engine.POST("/api/v1/add", func(ctx *gin.Context) { ctx.Status(tc.status) })

			req := httptest.NewRequest(http.MethodPost, "/api/v1/add", nil)
			req.Header.Set(RequestIDHeader, "req-1")
			req.Header.Set("User-Agent", "test-agent")
			serve(engine, req)

			require.Equal(t, 1, logs.Len())
			entry := logs.All()[0]
			assert.Equal(t, tc.level, entry.Level)
			assert.Equal(t, "http request", entry.Message)

			fields := entry.ContextMap()
			assert.Equal(t, "req-1", fields["request_id"])
			assert.Equal(t, "POST", fields["method"])
			assert.Equal(t, "/api/v1/add", fields["path"])
			assert.Equal(t, int64(tc.status), fields["status"])
			assert.Equal(t, "test-agent", fields["user_agent"])
			assert.Contains(t, fields, "latency")
			assert.Contains(t, fields, "client_ip")
			assert.NotContains(t, fields, "error")
		})
	}
}

func TestLoggerIncludesHandlerErrors(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	engine := gin.New()
	engine.Use(Logger(zap.New(core)))
	engine.GET("/", func(ctx *gin.Context) {
		_ = ctx.Error(errors.New("first"))
		_ = ctx.Error(errors.New("second"))
		ctx.Status(http.StatusUnprocessableEntity)
	})

	serve(engine, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, 1, logs.Len())
	assert.Equal(t, "first; second", logs.All()[0].ContextMap()["error"])
}

// --- Recovery ----------------------------------------------------------------

func TestRecoveryReturns500AndLogsPanic(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	log := zap.New(core)
	engine := gin.New()
	engine.Use(RequestID(), Logger(log), Recovery(log))
	engine.GET("/", func(*gin.Context) { panic("boom") })

	rec := serve(engine, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"INTERNAL_ERROR","message":"an unexpected error occurred"}}`, rec.Body.String())

	panics := logs.FilterMessage("panic recovered").All()
	require.Len(t, panics, 1)
	assert.Equal(t, zapcore.ErrorLevel, panics[0].Level)
	assert.Equal(t, "boom", panics[0].ContextMap()["panic"])
	assert.Contains(t, panics[0].ContextMap(), "stack")

	// The request line still records the 500 because Logger wraps Recovery.
	requests := logs.FilterMessage("http request").All()
	require.Len(t, requests, 1)
	assert.Equal(t, int64(http.StatusInternalServerError), requests[0].ContextMap()["status"])
}

// --- CORS --------------------------------------------------------------------

func newCORSEngine(origins ...string) *gin.Engine {
	engine := gin.New()
	engine.Use(CORS(origins))
	engine.POST("/api", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	return engine
}

func TestCORSAllowedOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := serve(newCORSEngine("http://localhost:5173"), req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://localhost:5173", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, RequestIDHeader, rec.Header().Get("Access-Control-Expose-Headers"))
	assert.Equal(t, "Origin", rec.Header().Get("Vary"))
}

func TestCORSWildcard(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api", nil)
	req.Header.Set("Origin", "http://anything.test")
	rec := serve(newCORSEngine("*"), req)

	assert.Equal(t, "http://anything.test", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSDisallowedOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api", nil)
	req.Header.Set("Origin", "http://evil.test")
	rec := serve(newCORSEngine("http://localhost:5173"), req)

	assert.Equal(t, http.StatusOK, rec.Code, "server still answers; the browser enforces CORS")
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSPreflight(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := serve(newCORSEngine("http://localhost:5173"), req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "GET, POST, OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}

// --- BodyLimit ---------------------------------------------------------------

func TestBodyLimit(t *testing.T) {
	engine := gin.New()
	engine.Use(BodyLimit(8))
	var readErr error
	engine.POST("/", func(ctx *gin.Context) { _, readErr = io.ReadAll(ctx.Request.Body) })

	serve(engine, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("12345678")))
	require.NoError(t, readErr)

	serve(engine, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("123456789")))
	var maxBytesErr *http.MaxBytesError
	assert.ErrorAs(t, readErr, &maxBytesErr)
}
