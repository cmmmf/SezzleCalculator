package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodyLimit caps the request body at maxBytes. Reading past the limit fails
// with *http.MaxBytesError, which the controller reports as 413.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxBytes)
		ctx.Next()
	}
}
