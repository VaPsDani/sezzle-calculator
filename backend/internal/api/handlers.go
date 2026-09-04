// Package api adapts the calculator domain to HTTP. It owns everything about
// the wire format: decoding requests, encoding responses, and choosing status
// codes. The calculator package knows nothing about any of it.
package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"

	"github.com/VaPsDani/sezzle-calculator/backend/internal/calculator"
)

// maxBodyBytes caps how much of a request body is read. Every valid request is
// a small JSON object with two numbers, so anything larger is either a mistake
// or an attempt to exhaust the server's memory.
const maxBodyBytes = 1 << 10 // 1 KiB

// calculationRequest is the body accepted by every operation endpoint.
//
// The fields are pointers on purpose. With a plain float64, the body {"a": 5}
// would leave B at its zero value and a divide request would fail with
// "division by zero", blaming the arithmetic for what is really a missing
// field. A nil pointer distinguishes "absent" from "sent as zero", which lets
// the handler answer 400 with an accurate reason.
type calculationRequest struct {
	A *float64 `json:"a"`
	B *float64 `json:"b"`
}

// calculationResponse is the body returned by a successful calculation.
// B is a pointer so that it is omitted entirely for unary operations such as
// sqrt, rather than reported as a meaningless zero.
type calculationResponse struct {
	Operation string   `json:"operation"`
	A         float64  `json:"a"`
	B         *float64 `json:"b,omitempty"`
	Result    float64  `json:"result"`
}

// binaryOperation is the shape every two operand operation is adapted to.
type binaryOperation func(a, b float64) (float64, error)

// unaryOperation is the shape of a single operand operation.
type unaryOperation func(a float64) (float64, error)

// operationHandlers returns the handler for each supported operation, keyed by
// the name used in both the route and the response body. The router ranges
// over this map, so adding an operation is a single line here.
func operationHandlers() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"add":        binaryHandler("add", total(calculator.Add)),
		"subtract":   binaryHandler("subtract", total(calculator.Subtract)),
		"multiply":   binaryHandler("multiply", total(calculator.Multiply)),
		"divide":     binaryHandler("divide", calculator.Divide),
		"power":      binaryHandler("power", calculator.Power),
		"percentage": binaryHandler("percentage", total(calculator.Percentage)),
		"sqrt":       unaryHandler("sqrt", calculator.Sqrt),
	}
}

// total adapts an operation that cannot fail to the binaryOperation shape.
//
// It also guards the result, which is not redundant here: JSON has no way to
// represent an infinity or a NaN, so encoding one fails after the status line
// has already been written and the client receives a truncated 200. Add,
// Subtract, Multiply and Percentage can all overflow to infinity with large
// enough operands, so the value is checked before it reaches the encoder.
func total(op func(a, b float64) float64) binaryOperation {
	return func(a, b float64) (float64, error) {
		result := op(a, b)
		if math.IsNaN(result) || math.IsInf(result, 0) {
			return 0, calculator.ErrResultOutOfRange
		}
		return result, nil
	}
}

// binaryHandler builds the handler for an operation that needs both operands.
// The handler shape is identical for every such operation, so the operation
// itself is a parameter instead of a copy of this function.
func binaryHandler(name string, op binaryOperation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeRequest(w, r)
		if err != nil {
			writeError(w, err)
			return
		}
		if req.A == nil {
			writeError(w, missingField("a"))
			return
		}
		if req.B == nil {
			writeError(w, missingField("b"))
			return
		}

		result, err := op(*req.A, *req.B)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, calculationResponse{
			Operation: name,
			A:         *req.A,
			B:         req.B,
			Result:    result,
		})
	}
}

// unaryHandler builds the handler for an operation that takes a single
// operand. A body carrying "b" is rejected rather than ignored, for the same
// reason the request fields are pointers: silently discarding input hides the
// caller's mistake instead of reporting it.
func unaryHandler(name string, op unaryOperation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeRequest(w, r)
		if err != nil {
			writeError(w, err)
			return
		}
		if req.A == nil {
			writeError(w, missingField("a"))
			return
		}
		if req.B != nil {
			writeError(w, unexpectedField("b", name))
			return
		}

		result, err := op(*req.A)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, calculationResponse{
			Operation: name,
			A:         *req.A,
			Result:    result,
		})
	}
}

// decodeRequest reads and validates the JSON body of a calculation request.
//
// It caps the number of bytes read, rejects fields the contract does not
// define, and refuses a body holding more than one JSON value. Every failure
// is translated into a requestError carrying the status and code to report.
func decodeRequest(w http.ResponseWriter, r *http.Request) (calculationRequest, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req calculationRequest
	if err := decoder.Decode(&req); err != nil {
		return calculationRequest{}, decodeError(err)
	}

	// A second value in the stream means the body was something like
	// {"a":1}{"a":2}. Only io.EOF confirms the body held exactly one object.
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return calculationRequest{}, trailingContent()
	}

	return req, nil
}

// writeJSON encodes payload as the response body with the given status.
//
// The status line is sent before encoding begins, so a failure here cannot be
// turned into an error response; all that is left is to record it.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("api: encoding %d response: %v", status, err)
	}
}
