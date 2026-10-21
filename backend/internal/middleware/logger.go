package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger writes one structured log line per request. The level follows the
// response status: 5xx → Error, 4xx → Warn, otherwise Info. Errors that
// handlers attached with ctx.Error are included, so each error is logged
// exactly once, here.
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		status := ctx.Writer.Status()
		fields := []zap.Field{
			zap.String("request_id", ctx.GetString(RequestIDKey)),
			zap.String("method", ctx.Request.Method),
			zap.String("path", ctx.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", ctx.ClientIP()),
			zap.String("user_agent", ctx.Request.UserAgent()),
		}
		if len(ctx.Errors) > 0 {
			fields = append(fields, zap.String("error", strings.Join(ctx.Errors.Errors(), "; ")))
		}

		log.Log(levelFor(status), "http request", fields...)
	}
}

func levelFor(status int) zapcore.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return zapcore.ErrorLevel
	case status >= http.StatusBadRequest:
		return zapcore.WarnLevel
	default:
		return zapcore.InfoLevel
	}
}
