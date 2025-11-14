package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusTeapot, response{"foo": "bar"})
	if w.Code != http.StatusTeapot {
		t.Errorf("status = %d; want %d", w.Code, http.StatusTeapot)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q; want %q", ct, "application/json")
	}
	body := strings.TrimSpace(w.Body.String())
	if body != `{"foo":"bar"}` {
		t.Errorf("body = %q; want %q", body, `{"foo":"bar"}`)
	}
}

func TestWriteJSONData(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	writeJSONData(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q; want %q", ct, "application/json")
	}

	body := strings.TrimSpace(w.Body.String())
	if !strings.Contains(body, `"key1":"value1"`) {
		t.Errorf("expected body to contain key1:value1, got %s", body)
	}
}
