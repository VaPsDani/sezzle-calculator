package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/VaPsDani/sezzle-calculator/backend/internal/calculator"
)

const unknownFieldPrefix = "json: unknown field "

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Carries its own status and code because both are only known where the
// problem is detected, while the detail is still available.
type requestError struct {
	status  int
	code    string
	message string
}

func (e *requestError) Error() string { return e.message }

func missingField(name string) error {
	return &requestError{
		status:  http.StatusBadRequest,
		code:    "MISSING_FIELD",
		message: fmt.Sprintf("field %q is required", name),
	}
}

func unexpectedField(name, operation string) error {
	return &requestError{
		status:  http.StatusBadRequest,
		code:    "UNEXPECTED_FIELD",
		message: fmt.Sprintf("field %q is not accepted by operation %q", name, operation),
	}
}

func unsupportedMediaType(received string) error {
	detail := "none was sent"
	if received != "" {
		detail = fmt.Sprintf("got %q", received)
	}
	return &requestError{
		status:  http.StatusUnsupportedMediaType,
		code:    "UNSUPPORTED_MEDIA_TYPE",
		message: fmt.Sprintf("content type must be application/json, %s", detail),
	}
}

func trailingContent() error {
	return &requestError{
		status:  http.StatusBadRequest,
		code:    "INVALID_JSON",
		message: "body must contain a single JSON object",
	}
}

func decodeError(err error) error {
	var (
		maxBytesErr *http.MaxBytesError
		syntaxErr   *json.SyntaxError
		typeErr     *json.UnmarshalTypeError
	)

	switch {
	case errors.As(err, &maxBytesErr):
		return &requestError{
			status:  http.StatusRequestEntityTooLarge,
			code:    "REQUEST_TOO_LARGE",
			message: fmt.Sprintf("request body must not exceed %d bytes", maxBytesErr.Limit),
		}

	case errors.As(err, &syntaxErr):
		return &requestError{
			status:  http.StatusBadRequest,
			code:    "INVALID_JSON",
			message: fmt.Sprintf("malformed JSON at byte %d", syntaxErr.Offset),
		}

	case errors.Is(err, io.ErrUnexpectedEOF):
		return &requestError{
			status:  http.StatusBadRequest,
			code:    "INVALID_JSON",
			message: "malformed JSON: unexpected end of input",
		}

	case errors.As(err, &typeErr):
		if typeErr.Field == "" {
			return &requestError{
				status:  http.StatusBadRequest,
				code:    "INVALID_JSON",
				message: "body must be a JSON object",
			}
		}
		// encoding/json reports an out of range number through the same error
		// as a wrong type, and only Value tells them apart.
		if strings.HasPrefix(typeErr.Value, "number ") {
			return &requestError{
				status:  http.StatusBadRequest,
				code:    "NUMBER_OUT_OF_RANGE",
				message: fmt.Sprintf("field %q is outside the range of a 64 bit float", typeErr.Field),
			}
		}
		return &requestError{
			status:  http.StatusBadRequest,
			code:    "INVALID_FIELD_TYPE",
			message: fmt.Sprintf("field %q must be a number", typeErr.Field),
		}

	case errors.Is(err, io.EOF):
		return &requestError{
			status:  http.StatusBadRequest,
			code:    "EMPTY_BODY",
			message: "request body is empty",
		}

	// encoding/json reports an unknown field as a plain errors.New, with no
	// exported type to match on, so this must come last to shadow nothing.
	case strings.HasPrefix(err.Error(), unknownFieldPrefix):
		return &requestError{
			status:  http.StatusBadRequest,
			code:    "UNKNOWN_FIELD",
			message: fmt.Sprintf("unknown field %s", strings.TrimPrefix(err.Error(), unknownFieldPrefix)),
		}

	default:
		return &requestError{
			status:  http.StatusBadRequest,
			code:    "INVALID_REQUEST",
			message: "request body could not be read",
		}
	}
}

func statusAndCode(err error) (int, string) {
	var reqErr *requestError
	if errors.As(err, &reqErr) {
		return reqErr.status, reqErr.code
	}

	switch {
	case errors.Is(err, calculator.ErrDivisionByZero):
		return http.StatusUnprocessableEntity, "DIVISION_BY_ZERO"

	case errors.Is(err, calculator.ErrNegativeSquareRoot):
		return http.StatusUnprocessableEntity, "NEGATIVE_SQUARE_ROOT"

	case errors.Is(err, calculator.ErrResultOutOfRange):
		return http.StatusUnprocessableEntity, "RESULT_OUT_OF_RANGE"

	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}

func writeError(w http.ResponseWriter, err error) {
	status, code := statusAndCode(err)

	message := err.Error()
	if status == http.StatusInternalServerError {
		// An unmapped error can carry internal detail, so it is logged, not sent.
		log.Printf("api: unmapped error: %v", err)
		message = "internal server error"
	}

	writeJSON(w, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}
