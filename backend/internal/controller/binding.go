package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/cmmmf/SezzleCalculator/backend/internal/model"
)

var errTrailingData = errors.New("request body must contain a single JSON object")

// bindStrictJSON decodes the body into dst, rejecting unknown fields and
// trailing data, then runs Gin's `binding` tag validation. Gin's own
// ShouldBindJSON only rejects unknown fields through a package-level flag, so
// we decode here to keep that behaviour local to this API.
func bindStrictJSON(ctx *gin.Context, dst any) error {
	dec := json.NewDecoder(ctx.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return errTrailingData
	}
	return binding.Validator.ValidateStruct(dst)
}

// respondBindError turns a decoding or validation error into a 400 (or 413)
// response with a message that tells the client what to fix.
func respondBindError(ctx *gin.Context, err error) {
	_ = ctx.Error(err).SetType(gin.ErrorTypeBind)

	status := http.StatusBadRequest
	var (
		maxBytesErr *http.MaxBytesError
		syntaxErr   *json.SyntaxError
		typeErr     *json.UnmarshalTypeError
		validErrs   validator.ValidationErrors
		message     string
	)
	switch {
	case errors.As(err, &maxBytesErr):
		status = http.StatusRequestEntityTooLarge
		message = fmt.Sprintf("request body must not exceed %d bytes", maxBytesErr.Limit)
	case errors.Is(err, io.EOF):
		message = "request body is required"
	case errors.Is(err, io.ErrUnexpectedEOF), errors.As(err, &syntaxErr):
		message = "request body must be valid JSON"
	case errors.As(err, &typeErr):
		if typeErr.Field == "" {
			message = "request body must be a JSON object"
		} else {
			message = fmt.Sprintf("field %q must be a valid number", typeErr.Field)
		}
	case errors.As(err, &validErrs):
		// JSON field names are the lower-cased struct field names (A → "a").
		message = fmt.Sprintf("field %q is required", strings.ToLower(validErrs[0].Field()))
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		message = "unknown field " + strings.TrimPrefix(err.Error(), "json: unknown field ")
	case errors.Is(err, errTrailingData):
		message = errTrailingData.Error()
	default:
		message = "invalid request body"
	}

	code := model.CodeInvalidRequest
	if status == http.StatusRequestEntityTooLarge {
		code = model.CodePayloadTooLarge
	}
	ctx.AbortWithStatusJSON(status, model.NewErrorResponse(code, message))
}
