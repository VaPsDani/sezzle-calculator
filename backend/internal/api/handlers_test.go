package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

const jsonContentType = "application/json; charset=utf-8"

func ptr(v float64) *float64 { return &v }

func TestRoutes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantAllow  string
		// Exactly one of the three below is set, and it decides how the body
		// is decoded and compared.
		wantResult *calculationResponse
		wantCode   string
		wantHealth *healthResponse
	}{
		{
			name:       "add returns the full response body",
			method:     http.MethodPost,
			path:       "/api/v1/add",
			body:       `{"a": 2, "b": 3}`,
			wantStatus: http.StatusOK,
			wantResult: &calculationResponse{Operation: "add", A: 2, B: ptr(3), Result: 5},
		},
		{
			name:       "divide returns a decimal result",
			method:     http.MethodPost,
			path:       "/api/v1/divide",
			body:       `{"a": 10, "b": 4}`,
			wantStatus: http.StatusOK,
			wantResult: &calculationResponse{Operation: "divide", A: 10, B: ptr(4), Result: 2.5},
		},
		{
			name:       "divide by zero is unprocessable",
			method:     http.MethodPost,
			path:       "/api/v1/divide",
			body:       `{"a": 10, "b": 0}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "DIVISION_BY_ZERO",
		},
		{
			name:       "square root of a negative is unprocessable",
			method:     http.MethodPost,
			path:       "/api/v1/sqrt",
			body:       `{"a": -9}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "NEGATIVE_SQUARE_ROOT",
		},
		{
			name:       "malformed JSON is rejected",
			method:     http.MethodPost,
			path:       "/api/v1/add",
			body:       `{no es json`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_JSON",
		},
		{
			name:       "missing operand is reported as such",
			method:     http.MethodPost,
			path:       "/api/v1/divide",
			body:       `{"a": 10}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "MISSING_FIELD",
		},
		{
			name:       "operand of the wrong type is rejected",
			method:     http.MethodPost,
			path:       "/api/v1/divide",
			body:       `{"a": 10, "b": "cuatro"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_FIELD_TYPE",
		},
		{
			name:       "GET on a POST only endpoint is not allowed",
			method:     http.MethodGet,
			path:       "/api/v1/divide",
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodPost,
			wantCode:   "METHOD_NOT_ALLOWED",
		},
		{
			name:       "health reports the service is up",
			method:     http.MethodGet,
			path:       "/api/v1/health",
			wantStatus: http.StatusOK,
			wantHealth: &healthResponse{Status: "ok"},
		},
	}

	router := NewRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.path, body)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if got := rec.Header().Get("Content-Type"); got != jsonContentType {
				t.Errorf("Content-Type = %q, want %q", got, jsonContentType)
			}
			if tt.wantAllow != "" {
				if got := rec.Header().Get("Allow"); got != tt.wantAllow {
					t.Errorf("Allow = %q, want %q", got, tt.wantAllow)
				}
			}

			switch {
			case tt.wantResult != nil:
				var got calculationResponse
				decodeBody(t, rec, &got)
				if !reflect.DeepEqual(got, *tt.wantResult) {
					t.Errorf("response = %+v, want %+v", got, *tt.wantResult)
				}

			case tt.wantHealth != nil:
				var got healthResponse
				decodeBody(t, rec, &got)
				if got != *tt.wantHealth {
					t.Errorf("response = %+v, want %+v", got, *tt.wantHealth)
				}

			case tt.wantCode != "":
				var got errorResponse
				decodeBody(t, rec, &got)
				if got.Error.Code != tt.wantCode {
					t.Errorf("error code = %q, want %q", got.Error.Code, tt.wantCode)
				}
				if got.Error.Message == "" {
					t.Error("error message is empty")
				}

			default:
				t.Fatal("test case declares no expected body")
			}
		})
	}
}

func TestUnaryOperationRejectsSecondOperand(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sqrt", strings.NewReader(`{"a": 9, "b": 2}`))
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var got errorResponse
	decodeBody(t, rec, &got)
	if got.Error.Code != "UNEXPECTED_FIELD" {
		t.Errorf("error code = %q, want %q", got.Error.Code, "UNEXPECTED_FIELD")
	}
}

func TestUnaryResponseOmitsSecondOperand(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sqrt", strings.NewReader(`{"a": 9}`))
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	var raw map[string]any
	decodeBody(t, rec, &raw)
	if _, present := raw["b"]; present {
		t.Errorf("response contains %q, want it omitted: %v", "b", raw)
	}
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decoding response body %q: %v", rec.Body.String(), err)
	}
}
