// Package router builds the Gin engine: middleware chain and routes.
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/cmmmf/SezzleCalculator/backend/internal/controller"
	"github.com/cmmmf/SezzleCalculator/backend/internal/middleware"
	"github.com/cmmmf/SezzleCalculator/backend/internal/model"
)

// MaxBodyBytes is the largest accepted request body. Calculator requests are
// a few dozen bytes, so 1 KB is generous.
const MaxBodyBytes = 1 << 10

// Options are the dependencies of the HTTP layer.
type Options struct {
	Logger      *zap.Logger
	CORSOrigins []string
	Calculator  *controller.CalculatorController
}

// New returns a Gin engine with the full middleware chain and all routes.
func New(opts Options) *gin.Engine {
	engine := gin.New()
	engine.HandleMethodNotAllowed = true
	// Never trust client-supplied X-Forwarded-For: client_ip is the TCP peer
	// (the nginx container when running under docker compose).
	_ = engine.SetTrustedProxies(nil)

	engine.Use(
		middleware.RequestID(),
		middleware.Logger(opts.Logger),
		middleware.Recovery(opts.Logger),
		middleware.CORS(opts.CORSOrigins),
		middleware.BodyLimit(MaxBodyBytes),
	)

	engine.GET("/healthz", controller.Health)

	calc := opts.Calculator
	v1 := engine.Group("/api/v1")
	v1.POST("/add", calc.Add)
	v1.POST("/subtract", calc.Subtract)
	v1.POST("/multiply", calc.Multiply)
	v1.POST("/divide", calc.Divide)
	v1.POST("/power", calc.Power)
	v1.POST("/percentage", calc.Percentage)
	v1.POST("/sqrt", calc.Sqrt)

	engine.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "route not found"))
	})
	engine.NoMethod(func(ctx *gin.Context) {
		ctx.JSON(http.StatusMethodNotAllowed,
			model.NewErrorResponse(model.CodeMethodNotAllowed, "method not allowed for this route"))
	})

	return engine
}
