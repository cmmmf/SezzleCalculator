package model

// Error codes are part of the API contract: clients branch on Code, while
// Message is meant for humans and may change.
const (
	CodeInvalidRequest     = "INVALID_REQUEST"
	CodeDivisionByZero     = "DIVISION_BY_ZERO"
	CodeNegativeSquareRoot = "NEGATIVE_SQUARE_ROOT"
	CodeResultOutOfRange   = "RESULT_OUT_OF_RANGE"
	CodeUndefinedResult    = "UNDEFINED_RESULT"
	CodeNotFound           = "NOT_FOUND"
	CodeMethodNotAllowed   = "METHOD_NOT_ALLOWED"
	CodePayloadTooLarge    = "PAYLOAD_TOO_LARGE"
	CodeInternalError      = "INTERNAL_ERROR"
)

// ErrorResponse is the body of every non-2xx response.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody describes what went wrong.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse builds an ErrorResponse.
func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{Error: ErrorBody{Code: code, Message: message}}
}
