package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"openclaw-honeypot/internal/logging"
	"openclaw-honeypot/internal/simulation"
)

func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request) {
	srcIP, srcPort := logging.SourceIP(r)

	// Log every HTTP request.
	attrs := []slog.Attr{
		slog.String("src_ip", srcIP),
		slog.String("src_port", srcPort),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("user_agent", r.UserAgent()),
		slog.String("host", r.Host),
	}

	// Read body for POST/PUT/PATCH.
	var body string
	if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
		b, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB limit
		if err == nil {
			body = string(b)
			attrs = append(attrs, slog.String("body", body))
		}
	}

	logging.Event(r.Context(), "http_request", attrs...)

	// Check for prompt injection in body.
	if body != "" {
		if detected, pattern := simulation.DetectPromptInjection(body); detected {
			logging.Event(r.Context(), "prompt_injection",
				slog.String("src_ip", srcIP),
				slog.String("path", r.URL.Path),
				slog.String("pattern", pattern),
				slog.String("raw_payload", body),
			)
		}
	}

	// Route.
	path := r.URL.Path
	switch {
	case r.Method == http.MethodPost && path == "/v1/chat/completions":
		s.handleChatCompletions(w, r, srcIP, body)
	case r.Method == http.MethodPost && path == "/tools/invoke":
		s.handleToolsInvoke(w, r, srcIP, body)
	default:
		// SPA shell for everything else.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Powered-By", "OpenClaw/0.14.2")
		fmt.Fprint(w, spaHTML)
	}
}

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request, srcIP, body string) {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		logging.Event(r.Context(), "ws_auth_attempt",
			slog.String("src_ip", srcIP),
			slog.String("auth_mode", "bearer"),
			slog.String("auth_value", strings.TrimPrefix(auth, "Bearer ")),
			slog.String("path", "/v1/chat/completions"),
		)
	}

	// Check Accept header for SSE.
	if strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		s.streamSSEResponse(w)
		return
	}

	// Check if request body asks for streaming.
	var req map[string]any
	if json.Unmarshal([]byte(body), &req) == nil {
		if stream, ok := req["stream"].(bool); ok && stream {
			s.streamSSEResponse(w)
			return
		}
	}

	// Non-streaming JSON response.
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"id":      "chatcmpl-" + simulation.NewSession().ID[:12],
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   "claude-sonnet-4-20250514",
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": "I'd be happy to help. Could you provide more details about what you're working on?",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]any{
			"prompt_tokens":     42,
			"completion_tokens": 18,
			"total_tokens":      60,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) streamSSEResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	chunks := []string{
		"I'd be", " happy to", " help.", " Could you", " provide more", " details?",
	}

	id := simulation.NewSession().ID[:12]
	for i, chunk := range chunks {
		data := map[string]any{
			"id":      "chatcmpl-" + id,
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   "claude-sonnet-4-20250514",
			"choices": []map[string]any{
				{
					"index": i,
					"delta": map[string]any{
						"content": chunk,
					},
				},
			},
		}
		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()
		time.Sleep(80 * time.Millisecond)
	}
	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (s *Server) handleToolsInvoke(w http.ResponseWriter, r *http.Request, srcIP, body string) {
	logging.Event(r.Context(), "ws_rpc_payload",
		slog.String("src_ip", srcIP),
		slog.String("method", "tools.invoke"),
		slog.String("raw_payload", body),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"result": map[string]any{
			"output":    "Tool executed successfully.",
			"exit_code": 0,
		},
	})
}
