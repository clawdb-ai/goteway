package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mac/goteway/internal/apperr"
	"github.com/mac/goteway/internal/observability"
)

// Server exposes compatibility HTTP routes.
type Server struct {
	logic  Logic
	health *observability.HealthReporter
}

// Logic defines the app methods required by HTTP compatibility endpoints.
type Logic interface {
	ChatCompletions(map[string]any) (map[string]any, error)
	InvokeTool(map[string]any) (map[string]any, error)
}

func New(logic Logic, health *observability.HealthReporter) *Server {
	return &Server{logic: logic, health: health}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/v1/chat/completions", s.handleCompletions)
	mux.HandleFunc("/tools/invoke", s.handleToolInvoke)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.health.Payload())
}

func (s *Server) handleCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	var in map[string]any
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, openAIError("invalid_request_error", "invalid json", "body", "invalid_json"))
		return
	}
	out, err := s.logic.ChatCompletions(in)
	if err != nil {
		if errors.Is(err, apperr.ErrNotImplemented) {
			writeJSON(w, http.StatusNotImplemented, openAIError("api_error", "chat completions not implemented", "", "not_implemented"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, openAIError("api_error", err.Error(), "", "internal_error"))
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleToolInvoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method not allowed"})
		return
	}
	var in map[string]any
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json"})
		return
	}
	out, err := s.logic.InvokeTool(in)
	if err != nil {
		if errors.Is(err, apperr.ErrNotImplemented) {
			writeJSON(w, http.StatusNotImplemented, map[string]any{"ok": false, "error": "tool invoke not implemented"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": out})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func openAIError(errType, message, param, code string) map[string]any {
	return map[string]any{
		"error": map[string]any{
			"message": message,
			"type":    errType,
			"param":   param,
			"code":    code,
		},
	}
}
