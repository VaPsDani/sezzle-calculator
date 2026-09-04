package api

import (
	"fmt"
	"net/http"
)

const apiPrefix = "/api/v1/"

type healthResponse struct {
	Status string `json:"status"`
}

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	methodsByPath := make(map[string]string)

	for name, handler := range operationHandlers() {
		path := apiPrefix + name
		mux.Handle(http.MethodPost+" "+path, handler)
		methodsByPath[path] = http.MethodPost
	}

	healthPath := apiPrefix + "health"
	mux.HandleFunc(http.MethodGet+" "+healthPath, handleHealth)
	methodsByPath[healthPath] = http.MethodGet

	mux.HandleFunc("/", fallbackHandler(methodsByPath))

	return mux
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

// The catch-all pattern matches whatever the specific ones did not, which
// includes the wrong verb on a real path, so it owns the 405 the mux would
// otherwise have produced.
func fallbackHandler(methodsByPath map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if method, ok := methodsByPath[r.URL.Path]; ok {
			w.Header().Set("Allow", method)
			writeError(w, &requestError{
				status:  http.StatusMethodNotAllowed,
				code:    "METHOD_NOT_ALLOWED",
				message: fmt.Sprintf("method %s is not allowed here, use %s", r.Method, method),
			})
			return
		}

		writeError(w, &requestError{
			status:  http.StatusNotFound,
			code:    "NOT_FOUND",
			message: "no such endpoint",
		})
	}
}
