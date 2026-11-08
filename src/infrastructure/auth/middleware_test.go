package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBearer(t *testing.T) {
	const token = "01234567890123456789012345678901"

	tests := []struct {
		name       string
		headers    []string
		statusCode int
		called     bool
	}{
		{name: "missing", statusCode: http.StatusUnauthorized},
		{name: "invalid", headers: []string{"Bearer invalid"}, statusCode: http.StatusUnauthorized},
		{name: "malformed", headers: []string{"Bearer"}, statusCode: http.StatusUnauthorized},
		{name: "empty", headers: []string{"Bearer    "}, statusCode: http.StatusUnauthorized},
		{name: "whitespace", headers: []string{"Bearer invalid token"}, statusCode: http.StatusUnauthorized},
		{name: "duplicate", headers: []string{"Bearer " + token, "Bearer " + token}, statusCode: http.StatusUnauthorized},
		{name: "valid", headers: []string{"bearer " + token}, statusCode: http.StatusNoContent, called: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			handler := Bearer(token, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusNoContent)
			}))

			req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
			for _, header := range test.headers {
				req.Header.Add("Authorization", header)
			}

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			if resp.Code != test.statusCode {
				t.Fatalf("expected status %d, got %d", test.statusCode, resp.Code)
			}
			if called != test.called {
				t.Fatalf("expected called=%v, got %v", test.called, called)
			}
			if !test.called && resp.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Fatal("missing WWW-Authenticate header")
			}
		})
	}
}
