// Package middleware holds cross-cutting Gin middleware: request IDs,
// structured request logging, panic recovery, CORS and body size limits.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"github.com/gin-gonic/gin"
)

const (
	// RequestIDHeader is read from requests and echoed on responses.
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the gin.Context key holding the request ID.
	RequestIDKey = "request_id"
)

// Incoming IDs are only trusted if they look like an ID; anything else (too
// long, or containing characters that could forge log lines) is replaced.
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// RequestID propagates the caller's X-Request-ID or generates a new one, so a
// request can be traced from the frontend to the backend logs.
func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.GetHeader(RequestIDHeader)
		if !validRequestID.MatchString(id) {
			id = newRequestID()
		}
		ctx.Set(RequestIDKey, id)
		ctx.Header(RequestIDHeader, id)
		ctx.Next()
	}
}

func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b) // crypto/rand.Read never returns an error.
	return hex.EncodeToString(b)
}
