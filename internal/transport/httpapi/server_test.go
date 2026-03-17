package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mac/goteway/internal/apperr"
	"github.com/mac/goteway/internal/observability"
)

type stubLogic struct {
	chatResp map[string]any
	chatErr  error
	toolResp map[string]any
	toolErr  error
}

func (s stubLogic) ChatCompletions(_ map[string]any) (map[string]any, error) {
	return s.chatResp, s.chatErr
}

func (s stubLogic) InvokeTool(_ map[string]any) (map[string]any, error) {
	return s.toolResp, s.toolErr
}

func TestHandleCompletionsMethodNotAllowed(t *testing.T) {
	srv := New(stubLogic{}, observability.NewHealthReporter("test"))
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if rec.Header().Get("Allow") != http.MethodPost {
		t.Fatalf("expected Allow header %q, got %q", http.MethodPost, rec.Header().Get("Allow"))
	}
}

func TestHandleCompletionsNotImplemented(t *testing.T) {
	srv := New(stubLogic{chatErr: apperr.ErrNotImplemented}, observability.NewHealthReporter("test"))
	h := srv.Handler()

	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"x"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if _, ok := out["error"]; !ok {
		t.Fatalf("expected error envelope, got %#v", out)
	}
}

func TestHandleCompletionsBodyTooLarge(t *testing.T) {
	srv := New(stubLogic{}, observability.NewHealthReporter("test"))
	h := srv.Handler()

	oversized := bytes.Repeat([]byte("a"), int(maxRequestBodyBytes)+1)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(oversized))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
