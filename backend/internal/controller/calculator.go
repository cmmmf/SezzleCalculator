// Package controller contains the Gin HTTP handlers. Controllers translate
// HTTP requests into service calls and service results into HTTP responses;
// they hold no business logic.
package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cmmmf/SezzleCalculator/backend/internal/model"
	"github.com/cmmmf/SezzleCalculator/backend/internal/service"
)

// Operation names, as returned in OperationResponse.Operation.
const (
	OpAdd        = "add"
	OpSubtract   = "subtract"
	OpMultiply   = "multiply"
	OpDivide     = "divide"
	OpPower      = "power"
	OpPercentage = "percentage"
	OpSqrt       = "sqrt"
)

// CalculatorController exposes CalculatorService over HTTP.
type CalculatorController struct {
	svc service.CalculatorService
}

// NewCalculatorController returns a controller backed by svc.
func NewCalculatorController(svc service.CalculatorService) *CalculatorController {
	return &CalculatorController{svc: svc}
}

// Add handles POST /api/v1/add.
func (c *CalculatorController) Add(ctx *gin.Context) { c.binary(ctx, OpAdd, c.svc.Add) }

// Subtract handles POST /api/v1/subtract.
func (c *CalculatorController) Subtract(ctx *gin.Context) {
	c.binary(ctx, OpSubtract, c.svc.Subtract)
}

// Multiply handles POST /api/v1/multiply.
func (c *CalculatorController) Multiply(ctx *gin.Context) {
	c.binary(ctx, OpMultiply, c.svc.Multiply)
}

// Divide handles POST /api/v1/divide.
func (c *CalculatorController) Divide(ctx *gin.Context) { c.binary(ctx, OpDivide, c.svc.Divide) }

// Power handles POST /api/v1/power.
func (c *CalculatorController) Power(ctx *gin.Context) { c.binary(ctx, OpPower, c.svc.Power) }

// Percentage handles POST /api/v1/percentage (a percent of b).
func (c *CalculatorController) Percentage(ctx *gin.Context) {
	c.binary(ctx, OpPercentage, c.svc.Percentage)
}

// Sqrt handles POST /api/v1/sqrt.
func (c *CalculatorController) Sqrt(ctx *gin.Context) {
	var req model.UnaryOperationRequest
	if err := bindStrictJSON(ctx, &req); err != nil {
		respondBindError(ctx, err)
		return
	}
	result, err := c.svc.Sqrt(*req.A)
	respond(ctx, OpSqrt, result, err)
}

// binary is the shared bind → call → respond flow of two-operand operations.
func (c *CalculatorController) binary(ctx *gin.Context, op string, fn func(a, b float64) (float64, error)) {
	var req model.BinaryOperationRequest
	if err := bindStrictJSON(ctx, &req); err != nil {
		respondBindError(ctx, err)
		return
	}
	result, err := fn(*req.A, *req.B)
	respond(ctx, op, result, err)
}

func respond(ctx *gin.Context, op string, result float64, err error) {
	if err != nil {
		respondServiceError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, model.OperationResponse{Operation: op, Result: result})
}

// serviceErrors maps domain errors to HTTP status codes and API error codes.
var serviceErrors = []struct {
	err    error
	status int
	code   string
}{
	{service.ErrInvalidOperand, http.StatusBadRequest, model.CodeInvalidRequest},
	{service.ErrDivisionByZero, http.StatusUnprocessableEntity, model.CodeDivisionByZero},
	{service.ErrNegativeSquareRoot, http.StatusUnprocessableEntity, model.CodeNegativeSquareRoot},
	{service.ErrResultOutOfRange, http.StatusUnprocessableEntity, model.CodeResultOutOfRange},
	{service.ErrUndefinedResult, http.StatusUnprocessableEntity, model.CodeUndefinedResult},
}

// respondServiceError writes the API error for err. The full error is
// attached to the context so the logging middleware records it once; the
// client only sees the domain error's own message, and nothing at all about
// unexpected errors.
func respondServiceError(ctx *gin.Context, err error) {
	_ = ctx.Error(err)

	for _, m := range serviceErrors {
		if errors.Is(err, m.err) {
			ctx.AbortWithStatusJSON(m.status, model.NewErrorResponse(m.code, m.err.Error()))
			return
		}
	}
	ctx.AbortWithStatusJSON(http.StatusInternalServerError,
		model.NewErrorResponse(model.CodeInternalError, "an unexpected error occurred"))
}
