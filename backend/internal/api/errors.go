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

// errorResponse is the body returned for every failed request.
type errorResponse struct {
	Error errorBody `json:"error"`
}

// errorBody carries a stable machine readable code plus a human readable
// message. Clients branch on Code; Message is only for people reading logs or
// screens, which leaves it free to be reworded or translated.
type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// requestError describes a request the server refuses to process before any
// arithmetic happens: malformed JSON, a missing operand, a wrong type. It
// carries its own status and code because those are decided at the point the
// problem is detected, where the detail is still available.
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

func trailingContent() error {
	return &requestError{
		status:  http.StatusBadRequest,
		code:    "INVALID_JSON",
		message: "body must contain a single JSON object",
	}
}

// decodeError translates a failure from encoding/json or from the body size
// limit into a requestError that names what the caller got wrong.
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
		// Field is empty when the mismatch is the body itself, as in a JSON
		// array where an object is expected.
		if typeErr.Field == "" {
			return &requestError{
				status:  http.StatusBadRequest,
				code:    "INVALID_JSON",
				message: "body must be a JSON object",
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
	// exported type or sentinel to match on, so the prefix is the only handle
	// available. It is checked last so that no typed case is shadowed by it.
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

// unknownFieldPrefix is the wording encoding/json uses when
// DisallowUnknownFields rejects a field.
const unknownFieldPrefix = "json: unknown field "

// statusAndCode maps an error to the HTTP status and the API error code that
// describe it. It is the single place where a domain error becomes a wire
// concern, which is what keeps the calculator package free of HTTP.
//
// Anything unrecognised is a bug in this server rather than in the request, so
// it maps to 500.
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

// writeError sends err to the client in the documented error envelope.
//
// The message of an unmapped error is replaced with a fixed string so that
// internal detail never reaches the client; the original is logged instead.
func writeError(w http.ResponseWriter, err error) {
	status, code := statusAndCode(err)

	message := err.Error()
	if status == http.StatusInternalServerError {
		log.Printf("api: unmapped error: %v", err)
		message = "internal server error"
	}

	writeJSON(w, status, errorResponse{Error: errorBody{Code: code, Message: message}})
}
