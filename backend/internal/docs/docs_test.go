package docs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func init() { gin.SetMode(gin.TestMode) }

func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	engine := gin.New()
	Register(engine)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestUIPageLoadsSpec(t *testing.T) {
	rec := get(t, UIPath)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), `url: "`+SpecPath+`"`)
	assert.Contains(t, rec.Body.String(), "swagger-ui-dist@"+swaggerUIVersion)
}

func TestSpecIsServedAndValidYAML(t *testing.T) {
	rec := get(t, SpecPath)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/yaml", rec.Header().Get("Content-Type"))

	var spec struct {
		OpenAPI string         `yaml:"openapi"`
		Paths   map[string]any `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(rec.Body.Bytes(), &spec))
	assert.Equal(t, "3.1.0", spec.OpenAPI)
	assert.NotEmpty(t, spec.Paths)
}
