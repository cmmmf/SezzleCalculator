package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// CORS allows browser calls from the given origins ("*" allows any). It is
// only needed when the frontend is served from a different origin; in Docker
// nginx proxies /api, so the browser sees a single origin.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowAll := slices.Contains(allowedOrigins, "*")
	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if origin != "" && (allowAll || slices.Contains(allowedOrigins, origin)) {
			h := ctx.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type, "+RequestIDHeader)
			h.Set("Access-Control-Expose-Headers", RequestIDHeader)
			h.Set("Access-Control-Max-Age", "600")
		}
		// Answer preflights here. For a disallowed origin the missing headers
		// make the browser block the real request.
		if ctx.Request.Method == http.MethodOptions && ctx.GetHeader("Access-Control-Request-Method") != "" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	}
}
