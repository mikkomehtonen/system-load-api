package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	Health(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want %q", body["status"], "ok")
	}
}

func TestCPU(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cpu", nil)
	w := httptest.NewRecorder()

	CPU(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if _, ok := body["timestamp"]; !ok {
		t.Error("response missing 'timestamp' field")
	}
	if _, ok := body["cpu"]; !ok {
		t.Error("response missing 'cpu' field")
	}
}

func TestMemory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/memory", nil)
	w := httptest.NewRecorder()

	Memory(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if _, ok := body["timestamp"]; !ok {
		t.Error("response missing 'timestamp' field")
	}
	if _, ok := body["memory"]; !ok {
		t.Error("response missing 'memory' field")
	}
}

func TestGPU(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/gpu", nil)
	w := httptest.NewRecorder()

	GPU(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if _, ok := body["timestamp"]; !ok {
		t.Error("response missing 'timestamp' field")
	}
	if _, ok := body["gpu"]; !ok {
		t.Error("response missing 'gpu' field")
	}

	gpu := body["gpu"].(map[string]interface{})
	if _, ok := gpu["available"]; !ok {
		t.Error("gpu object missing 'available' field")
	}
}

func TestDisk(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode (1s delta sampling)")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/disk", nil)
	w := httptest.NewRecorder()

	Disk(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if _, ok := body["timestamp"]; !ok {
		t.Error("response missing 'timestamp' field")
	}
	if _, ok := body["disk"]; !ok {
		t.Error("response missing 'disk' field")
	}
}

func TestNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode (1s delta sampling)")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/network", nil)
	w := httptest.NewRecorder()

	Network(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if _, ok := body["timestamp"]; !ok {
		t.Error("response missing 'timestamp' field")
	}
	if _, ok := body["network"]; !ok {
		t.Error("response missing 'network' field")
	}
}

func TestStats(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode (slow concurrent collection)")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	w := httptest.NewRecorder()

	Stats(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if _, ok := body["timestamp"]; !ok {
		t.Error("response missing 'timestamp' field")
	}
	if _, ok := body["cpu"]; !ok {
		t.Error("response missing 'cpu' field")
	}
	if _, ok := body["memory"]; !ok {
		t.Error("response missing 'memory' field")
	}
	if _, ok := body["disk"]; !ok {
		t.Error("response missing 'disk' field")
	}
	if _, ok := body["gpu"]; !ok {
		t.Error("response missing 'gpu' field")
	}
	if _, ok := body["network"]; !ok {
		t.Error("response missing 'network' field")
	}
}

func TestWriteJSON_SetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"hello": "world"})

	resp := w.Result()
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestWriteError_Returns500(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, "something broke")

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["error"] != "something broke" {
		t.Errorf("error = %q, want %q", body["error"], "something broke")
	}
}
