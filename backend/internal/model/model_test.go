package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin the JSON wire format, which is the contract with the frontend.

func TestErrorResponseJSON(t *testing.T) {
	body, err := json.Marshal(NewErrorResponse(CodeDivisionByZero, "division by zero is not allowed"))
	require.NoError(t, err)
	assert.JSONEq(t, `{"error":{"code":"DIVISION_BY_ZERO","message":"division by zero is not allowed"}}`, string(body))
}

func TestOperationResponseJSON(t *testing.T) {
	body, err := json.Marshal(OperationResponse{Operation: "divide", Result: 2.5})
	require.NoError(t, err)
	assert.JSONEq(t, `{"operation":"divide","result":2.5}`, string(body))
}

func TestBinaryOperationRequestDistinguishesMissingFromZero(t *testing.T) {
	var req BinaryOperationRequest
	require.NoError(t, json.Unmarshal([]byte(`{"a":0}`), &req))
	require.NotNil(t, req.A)
	assert.Equal(t, 0.0, *req.A)
	assert.Nil(t, req.B)
}
