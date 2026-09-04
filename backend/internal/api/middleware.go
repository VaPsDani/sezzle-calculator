package api

import (
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

const (
	defaultAllowedOrigin = "http://localhost:5173"
	preflightMaxAge      = "86400"
)

// The embedded ResponseWriter keeps every other method intact, so only the two
// that reveal the status are overridden.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}

// A handler that writes without calling WriteHeader gets an implicit 200.
func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	return r.ResponseWriter.Write(b)
}

func withCORS(next http.Handler) http.Handler {
	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" {
		origin = defaultAllowedOrigin
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", preflightMaxAge)
		w.Header().Add("Vary", "Origin")

		// The preflight only asks for permission, so it never reaches the mux,
		// which has no OPTIONS route and would answer 405.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(recorder, r)

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, recorder.status, time.Since(started).Round(time.Microsecond))
	})
}

func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		defer func() {
			panicked := recover()
			if panicked == nil {
				return
			}

			log.Printf("panic serving %s %s: %v\n%s", r.Method, r.URL.Path, panicked, debug.Stack())

			// Half of the response may already be on the wire, in which case
			// the status is set and there is nothing left to correct.
			if !recorder.wroteHeader {
				writeJSON(recorder, http.StatusInternalServerError, errorResponse{
					Error: errorBody{Code: "INTERNAL_ERROR", Message: "internal server error"},
				})
			}
		}()

		next.ServeHTTP(recorder, r)
	})
}
