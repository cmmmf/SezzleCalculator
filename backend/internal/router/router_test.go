package router

import (
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

	"github.com/cmmmf/SezzleCalculator/backend/internal/controller"
	"github.com/cmmmf/SezzleCalculator/backend/internal/middleware"
	"github.com/cmmmf/SezzleCalculator/backend/internal/service"
)

// These are end-to-end tests of the HTTP layer with the real service: they
// check that routes, middleware and error handling are wired together.

func init() { gin.SetMode(gin.TestMode) }

func newEngine(t *testing.T) (*gin.Engine, *observer.ObservedLogs) {
	t.Helper()
	core, logs := observer.New(zapcore.DebugLevel)
	return New(Options{
		Logger:      zap.New(core),
		CORSOrigins: []string{"http://localhost:5173"},
		Calculator:  controller.NewCalculatorController(service.NewCalculatorService()),
	}), logs
}

func post(engine http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestCalculatorEndpoints(t *testing.T) {
	engine, _ := newEngine(t)
	cases := []struct {
		path, body, want string
		status           int
	}{
		{"/api/v1/add", `{"a":10,"b":4}`, `{"operation":"add","result":14}`, 200},
		{"/api/v1/subtract", `{"a":10,"b":4}`, `{"operation":"subtract","result":6}`, 200},
		{"/api/v1/multiply", `{"a":10,"b":4}`, `{"operation":"multiply","result":40}`, 200},
		{"/api/v1/divide", `{"a":10,"b":4}`, `{"operation":"divide","result":2.5}`, 200},
		{"/api/v1/power", `{"a":2,"b":10}`, `{"operation":"power","result":1024}`, 200},
		{"/api/v1/percentage", `{"a":15,"b":200}`, `{"operation":"percentage","result":30}`, 200},
		{"/api/v1/sqrt", `{"a":16}`, `{"operation":"sqrt","result":4}`, 200},
		{"/api/v1/divide", `{"a":10,"b":0}`, `{"error":{"code":"DIVISION_BY_ZERO","message":"division by zero is not allowed"}}`, 422},
		{"/api/v1/sqrt", `{"a":-4}`, `{"error":{"code":"NEGATIVE_SQUARE_ROOT","message":"square root of a negative number is not a real number"}}`, 422},
		{"/api/v1/power", `{"a":10,"b":400}`, `{"error":{"code":"RESULT_OUT_OF_RANGE","message":"result is too large to be represented"}}`, 422},
		{"/api/v1/power", `{"a":-8,"b":0.5}`, `{"error":{"code":"UNDEFINED_RESULT","message":"result is undefined for the given operands"}}`, 422},
		{"/api/v1/add", `{"a":1}`, `{"error":{"code":"INVALID_REQUEST","message":"field \"b\" is required"}}`, 400},
	}
	for _, tc := range cases {
		t.Run(tc.path+" "+tc.body, func(t *testing.T) {
			rec := post(engine, tc.path, tc.body)
			assert.Equal(t, tc.status, rec.Code)
			assert.JSONEq(t, tc.want, rec.Body.String())
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		})
	}
}

func TestHealthz(t *testing.T) {
	engine, _ := newEngine(t)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func TestNotFound(t *testing.T) {
	engine, _ := newEngine(t)
	rec := post(engine, "/api/v1/modulo", `{"a":1,"b":2}`)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"NOT_FOUND","message":"route not found"}}`, rec.Body.String())
}

func TestMethodNotAllowed(t *testing.T) {
	engine, _ := newEngine(t)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/add", nil))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"METHOD_NOT_ALLOWED","message":"method not allowed for this route"}}`, rec.Body.String())
}

func TestPayloadTooLarge(t *testing.T) {
	engine, _ := newEngine(t)
	body := `{"a":1,"b":2,"pad":"` + strings.Repeat("x", MaxBodyBytes) + `"}`
	rec := post(engine, "/api/v1/add", body)
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.Contains(t, rec.Body.String(), `"code":"PAYLOAD_TOO_LARGE"`)
}

func TestMiddlewareChain(t *testing.T) {
	engine, logs := newEngine(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/divide", strings.NewReader(`{"a":1,"b":0}`))
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set(middleware.RequestIDHeader, "trace-42")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	assert.Equal(t, "trace-42", rec.Header().Get(middleware.RequestIDHeader))
	assert.Equal(t, "http://localhost:5173", rec.Header().Get("Access-Control-Allow-Origin"))

	require.Equal(t, 1, logs.Len(), "each request is logged exactly once")
	entry := logs.All()[0]
	assert.Equal(t, zapcore.WarnLevel, entry.Level)
	assert.Equal(t, "trace-42", entry.ContextMap()["request_id"])
	assert.Contains(t, entry.ContextMap()["error"], "division by zero")
}

func TestCORSPreflightOnAPIRoute(t *testing.T) {
	engine, _ := newEngine(t)
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/add", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "http://localhost:5173", rec.Header().Get("Access-Control-Allow-Origin"))
}
