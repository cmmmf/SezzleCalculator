package controller

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cmmmf/SezzleCalculator/backend/internal/service"
)

// fakeService records which method was called with which operands and
// returns a canned result, so controller tests don't depend on the real math.
type fakeService struct {
	result float64
	err    error

	method   string
	operands []float64
}

func (f *fakeService) call(method string, operands ...float64) (float64, error) {
	f.method, f.operands = method, operands
	return f.result, f.err
}

func (f *fakeService) Add(a, b float64) (float64, error)      { return f.call("Add", a, b) }
func (f *fakeService) Subtract(a, b float64) (float64, error) { return f.call("Subtract", a, b) }
func (f *fakeService) Multiply(a, b float64) (float64, error) { return f.call("Multiply", a, b) }
func (f *fakeService) Divide(a, b float64) (float64, error)   { return f.call("Divide", a, b) }
func (f *fakeService) Power(a, b float64) (float64, error)    { return f.call("Power", a, b) }
func (f *fakeService) Percentage(a, b float64) (float64, error) {
	return f.call("Percentage", a, b)
}
func (f *fakeService) Sqrt(a float64) (float64, error) { return f.call("Sqrt", a) }

func init() { gin.SetMode(gin.TestMode) }

func newTestEngine(svc service.CalculatorService) *gin.Engine {
	c := NewCalculatorController(svc)
	engine := gin.New()
	engine.GET("/healthz", Health)
	engine.POST("/add", c.Add)
	engine.POST("/subtract", c.Subtract)
	engine.POST("/multiply", c.Multiply)
	engine.POST("/divide", c.Divide)
	engine.POST("/power", c.Power)
	engine.POST("/percentage", c.Percentage)
	engine.POST("/sqrt", c.Sqrt)
	return engine
}

func do(engine http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestRoutesCallMatchingServiceMethod(t *testing.T) {
	cases := []struct {
		path, body, method, operation string
		operands                      []float64
	}{
		{"/add", `{"a":10,"b":4}`, "Add", OpAdd, []float64{10, 4}},
		{"/subtract", `{"a":10,"b":4}`, "Subtract", OpSubtract, []float64{10, 4}},
		{"/multiply", `{"a":10,"b":4}`, "Multiply", OpMultiply, []float64{10, 4}},
		{"/divide", `{"a":10,"b":4}`, "Divide", OpDivide, []float64{10, 4}},
		{"/power", `{"a":2,"b":10}`, "Power", OpPower, []float64{2, 10}},
		{"/percentage", `{"a":15,"b":200}`, "Percentage", OpPercentage, []float64{15, 200}},
		{"/sqrt", `{"a":16}`, "Sqrt", OpSqrt, []float64{16}},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			svc := &fakeService{result: 42.5}
			rec := do(newTestEngine(svc), http.MethodPost, tc.path, strings.NewReader(tc.body))

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, `{"operation":"`+tc.operation+`","result":42.5}`, rec.Body.String())
			assert.Equal(t, tc.method, svc.method)
			assert.Equal(t, tc.operands, svc.operands)
		})
	}
}

func TestZeroOperandIsAccepted(t *testing.T) {
	svc := &fakeService{}
	rec := do(newTestEngine(svc), http.MethodPost, "/add", strings.NewReader(`{"a":0,"b":0}`))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []float64{0, 0}, svc.operands)
}

func TestServiceErrorMapping(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		status  int
		code    string
		message string
	}{
		{"division by zero", service.ErrDivisionByZero, http.StatusUnprocessableEntity, "DIVISION_BY_ZERO", service.ErrDivisionByZero.Error()},
		{"negative sqrt", service.ErrNegativeSquareRoot, http.StatusUnprocessableEntity, "NEGATIVE_SQUARE_ROOT", service.ErrNegativeSquareRoot.Error()},
		{"out of range", service.ErrResultOutOfRange, http.StatusUnprocessableEntity, "RESULT_OUT_OF_RANGE", service.ErrResultOutOfRange.Error()},
		{"undefined", service.ErrUndefinedResult, http.StatusUnprocessableEntity, "UNDEFINED_RESULT", service.ErrUndefinedResult.Error()},
		{"invalid operand", service.ErrInvalidOperand, http.StatusBadRequest, "INVALID_REQUEST", service.ErrInvalidOperand.Error()},
		{"wrapped domain error exposes only the domain message", errors.Join(errors.New("internal context"), service.ErrDivisionByZero), http.StatusUnprocessableEntity, "DIVISION_BY_ZERO", service.ErrDivisionByZero.Error()},
		{"unexpected error is hidden", errors.New("db password is hunter2"), http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeService{err: tc.err}
			rec := do(newTestEngine(svc), http.MethodPost, "/divide", strings.NewReader(`{"a":1,"b":0}`))

			assert.Equal(t, tc.status, rec.Code)
			assert.JSONEq(t,
				`{"error":{"code":"`+tc.code+`","message":`+strconv.Quote(tc.message)+`}}`,
				rec.Body.String())
		})
	}
}

func TestSqrtServiceError(t *testing.T) {
	svc := &fakeService{err: service.ErrNegativeSquareRoot}
	rec := do(newTestEngine(svc), http.MethodPost, "/sqrt", strings.NewReader(`{"a":-4}`))
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "NEGATIVE_SQUARE_ROOT")
}

func TestInvalidRequestBodies(t *testing.T) {
	cases := []struct {
		name, path, body, message string
	}{
		{"empty body", "/add", ``, "request body is required"},
		{"malformed JSON", "/add", `{"a":1,`, "request body must be valid JSON"},
		{"syntax error", "/add", `{"a":1 "b":2}`, "request body must be valid JSON"},
		{"not an object", "/add", `[1,2]`, "request body must be a JSON object"},
		{"missing b", "/add", `{"a":1}`, `field "b" is required`},
		{"missing a", "/divide", `{"b":1}`, `field "a" is required`},
		{"null operand", "/add", `{"a":null,"b":1}`, `field "a" is required`},
		{"string operand", "/add", `{"a":"1","b":2}`, `field "a" must be a valid number`},
		{"number out of float64 range", "/multiply", `{"a":1e400,"b":2}`, `field "a" must be a valid number`},
		{"unknown field", "/add", `{"a":1,"b":2,"c":3}`, `unknown field "c"`},
		{"sqrt with extra operand", "/sqrt", `{"a":4,"b":2}`, `unknown field "b"`},
		{"sqrt missing operand", "/sqrt", `{}`, `field "a" is required`},
		{"trailing data", "/add", `{"a":1,"b":2}{"a":3}`, "request body must contain a single JSON object"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeService{}
			rec := do(newTestEngine(svc), http.MethodPost, tc.path, strings.NewReader(tc.body))

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.JSONEq(t,
				`{"error":{"code":"INVALID_REQUEST","message":`+strconv.Quote(tc.message)+`}}`,
				rec.Body.String())
			assert.Empty(t, svc.method, "service must not be called for invalid input")
		})
	}
}

func TestBindErrorsAreAttachedForLogging(t *testing.T) {
	engine := gin.New()
	var captured []*gin.Error
	engine.Use(func(ctx *gin.Context) {
		ctx.Next()
		captured = ctx.Errors
	})
	engine.POST("/add", NewCalculatorController(&fakeService{}).Add)

	do(engine, http.MethodPost, "/add", strings.NewReader(`{"a":1}`))
	require.Len(t, captured, 1)
	assert.Equal(t, gin.ErrorTypeBind, captured[0].Type)
}

func TestHealth(t *testing.T) {
	rec := do(newTestEngine(&fakeService{}), http.MethodGet, "/healthz", nil)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func TestPayloadTooLarge(t *testing.T) {
	engine := gin.New()
	engine.Use(func(ctx *gin.Context) {
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 8)
	})
	engine.POST("/add", NewCalculatorController(&fakeService{}).Add)

	rec := do(engine, http.MethodPost, "/add", strings.NewReader(`{"a":1,"b":2}`))

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"PAYLOAD_TOO_LARGE","message":"request body must not exceed 8 bytes"}}`, rec.Body.String())
}
