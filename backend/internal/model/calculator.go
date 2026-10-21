// Package model defines the JSON request and response shapes of the REST API.
package model

// Operands are pointers so that a missing field fails `binding:"required"`
// instead of silently defaulting to 0.

// BinaryOperationRequest is the body of operations that take two operands.
type BinaryOperationRequest struct {
	A *float64 `json:"a" binding:"required"`
	B *float64 `json:"b" binding:"required"`
}

// UnaryOperationRequest is the body of operations that take one operand.
type UnaryOperationRequest struct {
	A *float64 `json:"a" binding:"required"`
}

// OperationResponse is returned when an operation succeeds.
type OperationResponse struct {
	Operation string  `json:"operation"`
	Result    float64 `json:"result"`
}

// HealthResponse is returned by the health check endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}
