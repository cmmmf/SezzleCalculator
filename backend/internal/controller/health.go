package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cmmmf/SezzleCalculator/backend/internal/model"
)

// Health handles GET /healthz. The service has no dependencies, so being able
// to answer means it is healthy.
func Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, model.HealthResponse{Status: "ok"})
}
