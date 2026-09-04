// Package api adapts the calculator domain to HTTP.
package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"mime"
	"net/http"

	"github.com/VaPsDani/sezzle-calculator/backend/internal/calculator"
)

const maxBodyBytes = 1 << 10

// The fields are pointers so that a missing operand is nil instead of 0:
// {"a": 5} must answer 400, not "division by zero".
type calculationRequest struct {
	A *float64 `json:"a"`
	B *float64 `json:"b"`
}

type calculationResponse struct {
	Operation string   `json:"operation"`
	A         float64  `json:"a"`
	B         *float64 `json:"b,omitempty"`
	Result    float64  `json:"result"`
}

type binaryOperation func(a, b float64) (float64, error)

type unaryOperation func(a float64) (float64, error)

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

// JSON cannot encode Inf or NaN, and by then the status line is already sent,
// so operations that cannot report overflow themselves are guarded here.
func total(op func(a, b float64) float64) binaryOperation {
	return func(a, b float64) (float64, error) {
		result := op(a, b)
		if math.IsNaN(result) || math.IsInf(result, 0) {
			return 0, calculator.ErrResultOutOfRange
		}
		return result, nil
	}
}

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

func decodeRequest(w http.ResponseWriter, r *http.Request) (calculationRequest, error) {
	if err := requireJSONContentType(r); err != nil {
		return calculationRequest{}, err
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req calculationRequest
	if err := decoder.Decode(&req); err != nil {
		return calculationRequest{}, decodeError(err)
	}

	// Only io.EOF confirms the body held a single object; anything else means
	// trailing content such as {"a":1}{"a":2}, which Decode would ignore.
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return calculationRequest{}, trailingContent()
	}

	return req, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("api: encoding %d response: %v", status, err)
	}
}

// Without this the server guesses the format from the bytes, so a client that
// sends XML gets told its JSON is malformed.
func requireJSONContentType(r *http.Request) error {
	header := r.Header.Get("Content-Type")
	if header == "" {
		return unsupportedMediaType("")
	}

	mediaType, _, err := mime.ParseMediaType(header)
	if err != nil || mediaType != "application/json" {
		return unsupportedMediaType(header)
	}
	return nil
}
