package main

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetPort_Default(t *testing.T) {
	os.Unsetenv("PORT")
	port := getPort()
	if port != "8080" {
		t.Errorf("getPort() = %q, want %q", port, "8080")
	}
}

func TestGetPort_Env(t *testing.T) {
	os.Setenv("PORT", "9090")
	defer os.Unsetenv("PORT")
	port := getPort()
	if port != "9090" {
		t.Errorf("getPort() = %q, want %q", port, "9090")
	}
}

func TestRoutesRegistered(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode (slow concurrent collection)")
	}

	staticSub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		t.Fatalf("failed to sub static: %v", err)
	}

	handler := buildHandler(staticSub)

	paths := []string{
		"/health",
		"/api/v1/stats",
		"/api/v1/cpu",
		"/api/v1/memory",
		"/api/v1/disk",
		"/api/v1/gpu",
		"/api/v1/network",
		"/api/v1/host",
	}

	for _, p := range paths {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", p, w.Code, http.StatusOK)
			continue
		}

		if p == "/health" {
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("GET %s decode error: %v", p, err)
			}
			if body["status"] != "ok" {
				t.Errorf("GET %s status = %q, want %q", p, body["status"], "ok")
			}
			continue
		}

		var body map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("GET %s decode error: %v", p, err)
		}
		if _, ok := body["timestamp"]; !ok {
			t.Errorf("GET %s missing timestamp", p)
		}
	}
}

func TestStaticServed(t *testing.T) {
	staticSub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		t.Fatalf("failed to sub static: %v", err)
	}

	handler := buildHandler(staticSub)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET / status = %d, want %d", w.Code, http.StatusOK)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" && ct != "text/html" {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}
