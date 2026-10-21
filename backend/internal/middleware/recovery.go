package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/cmmmf/SezzleCalculator/backend/internal/model"
)

// Recovery turns a panic into a 500 INTERNAL_ERROR response and logs it with
// its stack trace. Register it after Logger so the request line records the
// 500 status.
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.String("request_id", ctx.GetString(RequestIDKey)),
					zap.Any("panic", r),
					zap.Stack("stack"),
				)
				ctx.AbortWithStatusJSON(http.StatusInternalServerError,
					model.NewErrorResponse(model.CodeInternalError, "an unexpected error occurred"))
			}
		}()
		ctx.Next()
	}
}
