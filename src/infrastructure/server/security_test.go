package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cgouz/pgo-mcp/src/infrastructure/config"
)

func TestMCPRequestBodyLimit(t *testing.T) {
	const token = "01234567890123456789012345678901"

	server := New("127.0.0.1", "8002", nil, token, config.Config{})
	body := strings.NewReader(strings.Repeat("x", maxRequestBodyBytes+1))
	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp := httptest.NewRecorder()
	server.http.Handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, resp.Code)
	}
}

func TestHTTPServerLimits(t *testing.T) {
	server := New("127.0.0.1", "8002", nil, "token", config.Config{})

	if server.http.ReadHeaderTimeout <= 0 {
		t.Fatal("ReadHeaderTimeout must be configured")
	}
	if server.http.IdleTimeout <= 0 {
		t.Fatal("IdleTimeout must be configured")
	}
	if server.http.MaxHeaderBytes != maxHeaderBytes {
		t.Fatalf("expected MaxHeaderBytes %d, got %d", maxHeaderBytes, server.http.MaxHeaderBytes)
	}
}
