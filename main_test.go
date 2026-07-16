package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConnectHandlerRejectsDisallowedTarget(t *testing.T) {
	handler := connectHandler("10.43.0.1:443")

	req := httptest.NewRequest(http.MethodConnect, "http://evil.example.com:443", nil)
	req.Host = "evil.example.com:443"
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for disallowed target, got %d", rec.Code)
	}
}

func TestConnectHandlerRejectsNonConnectMethod(t *testing.T) {
	handler := connectHandler("10.43.0.1:443")

	req := httptest.NewRequest(http.MethodGet, "http://10.43.0.1:443", nil)
	req.Host = "10.43.0.1:443"
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-CONNECT method, got %d", rec.Code)
	}
}
