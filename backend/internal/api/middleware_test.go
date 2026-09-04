package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// silenceLogs keeps the expected panic stack and error lines out of the test
// output, and restores the logger when the test ends.
func silenceLogs(t *testing.T) {
	t.Helper()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
}

func TestPreflightIsAnsweredWithoutReachingTheMux(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/divide", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)

	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, req)

	// The mux has no OPTIONS route, so anything other than 204 means the
	// request went past the CORS middleware and was answered as a 405.
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	headers := map[string]string{
		"Access-Control-Allow-Origin":  "http://localhost:5173",
		"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type",
	}
	for name, want := range headers {
		if got := rec.Header().Get(name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestCORSHeadersAreSentOnANormalResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, newJSONRequest(http.MethodPost, "/api/v1/add", `{"a":1,"b":2}`))

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "" {
		t.Error("Access-Control-Allow-Origin is missing from a normal response")
	}
}

func TestRecoveryTurnsAPanicIntoAnInternalServerError(t *testing.T) {
	silenceLogs(t)

	exploding := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("a bug that reached production")
	})
	chain := withLogging(withRecovery(withCORS(exploding)))

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, newJSONRequest(http.MethodPost, "/api/v1/add", `{"a":1,"b":2}`))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var got errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decoding %q: %v", rec.Body.String(), err)
	}
	if got.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("error code = %q, want %q", got.Error.Code, "INTERNAL_ERROR")
	}
	if strings.Contains(got.Error.Message, "bug") {
		t.Errorf("message leaks the panic value: %q", got.Error.Message)
	}
}

func TestLoggingRecordsTheStatusWrittenByAHandler(t *testing.T) {
	var recorded strings.Builder
	log.SetOutput(&recorded)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	rec := httptest.NewRecorder()
	NewRouter().ServeHTTP(rec, newJSONRequest(http.MethodPost, "/api/v1/divide", `{"a":10,"b":0}`))

	if line := recorded.String(); !strings.Contains(line, "422") {
		t.Errorf("access log = %q, want it to contain the 422 status", line)
	}
}

func TestUnmappedErrorBecomesInternalServerErrorWithoutLeaking(t *testing.T) {
	silenceLogs(t)

	rec := httptest.NewRecorder()
	writeError(rec, errors.New("connection string: user:password@db"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Errorf("response leaks internal detail: %s", rec.Body.String())
	}
}

func TestDecodeErrorFallsBackToInvalidRequest(t *testing.T) {
	var reqErr *requestError
	if !errors.As(decodeError(errors.New("something no case matches")), &reqErr) {
		t.Fatal("decodeError did not return a requestError")
	}
	if reqErr.status != http.StatusBadRequest || reqErr.code != "INVALID_REQUEST" {
		t.Errorf("got %d %s, want 400 INVALID_REQUEST", reqErr.status, reqErr.code)
	}
}

func TestWriteJSONDoesNotPanicOnAValueItCannotEncode(t *testing.T) {
	silenceLogs(t)

	rec := httptest.NewRecorder()
	// JSON has no representation for an infinity, so the encoder fails after
	// the status line is already on the wire.
	writeJSON(rec, http.StatusOK, map[string]float64{"result": math.Inf(1)})

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
