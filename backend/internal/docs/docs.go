// Package docs serves the OpenAPI specification of the API and a Swagger UI
// page that renders it.
//
// The spec is written by hand and embedded in the binary, so it needs no code
// generator and no extra dependencies. The router tests check that it lists
// exactly the routes the server registers.
package docs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Paths under which the docs are served.
const (
	UIPath   = "/swagger"
	SpecPath = "/swagger/openapi.yaml"
)

// Spec is the OpenAPI 3.1 document of the API.
//
//go:embed openapi.yaml
var Spec []byte

// swaggerUIVersion pins the swagger-ui-dist release loaded from the CDN.
const swaggerUIVersion = "5.17.14"

const uiPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Calculator API · Swagger UI</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({ url: "` + SpecPath + `", dom_id: "#swagger-ui" });
  </script>
</body>
</html>`

// Register adds the Swagger UI page and the raw spec to r.
func Register(r gin.IRoutes) {
	r.GET(UIPath, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(uiPage))
	})
	r.GET(SpecPath, func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "application/yaml", Spec)
	})
}
