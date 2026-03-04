package simulation

import (
	"encoding/json"
	"strings"
)

// RPCResponse returns a plausible fake response for a given RPC method.
func RPCResponse(method string, params json.RawMessage) map[string]any {
	switch {
	case method == "health":
		return map[string]any{
			"status": "ok",
			"uptime": 86400,
			"version": map[string]any{
				"server":   "0.14.2",
				"protocol": 3,
				"runtime":  "node/22.11.0",
			},
		}
	case method == "chat.send" || method == "chat.message":
		return map[string]any{
			"id":      generateID(),
			"role":    "assistant",
			"content": "I'm sorry, I can't help with that request. Please try rephrasing your question.",
			"model":   "claude-sonnet-4-20250514",
			"usage":   map[string]any{"input_tokens": 142, "output_tokens": 23},
		}
	case method == "agent.status":
		return map[string]any{
			"id":     "agent-primary",
			"status": "idle",
			"model":  "claude-sonnet-4-20250514",
			"capabilities": []string{
				"chat", "tools", "browser", "shell", "code_edit",
				"file_read", "file_write", "search",
			},
			"version":    "0.14.2",
			"max_tokens": 8192,
		}
	case method == "config.get":
		return map[string]any{
			"server_name":    "openclaw-prod",
			"listen_address": "0.0.0.0:18789",
			"auth_mode":      "token",
			"tls_enabled":    false,
			"model_provider": "anthropic",
			"model_name":     "claude-sonnet-4-20250514",
			"api_key":        "sk-ant-...REDACTED",
			"max_sessions":   50,
			"log_level":      "info",
			"tools_enabled":  true,
			"sandbox_mode":   "docker",
		}
	case method == "tools.catalog" || method == "tools.list":
		return map[string]any{
			"tools": []map[string]any{
				{"name": "browser", "version": "1.2.0", "enabled": true},
				{"name": "shell", "version": "1.0.3", "enabled": true},
				{"name": "code_edit", "version": "1.1.0", "enabled": true},
				{"name": "file_read", "version": "1.0.0", "enabled": true},
				{"name": "file_write", "version": "1.0.0", "enabled": true},
				{"name": "search", "version": "0.9.1", "enabled": true},
				{"name": "email", "version": "0.8.0", "enabled": false},
				{"name": "calendar", "version": "0.7.2", "enabled": false},
			},
		}
	case method == "nodes.list":
		return map[string]any{
			"nodes": []map[string]any{
				{
					"id": "node-01", "hostname": "openclaw-prod-01",
					"status": "healthy", "uptime": 604800,
					"cpu_usage": 0.23, "mem_usage": 0.41,
				},
			},
		}
	case method == "logs.tail":
		return map[string]any{
			"entries": []map[string]any{
				{"ts": "2026-03-04T11:59:58Z", "level": "info", "msg": "request completed", "duration_ms": 142},
				{"ts": "2026-03-04T11:59:59Z", "level": "info", "msg": "health check ok"},
				{"ts": "2026-03-04T12:00:00Z", "level": "debug", "msg": "tick sent to 3 clients"},
			},
		}
	case strings.HasPrefix(method, "tools.invoke"):
		return map[string]any{
			"status": "ok",
			"result": map[string]any{
				"output":    "Command executed successfully.",
				"exit_code": 0,
			},
		}
	case strings.HasPrefix(method, "config."):
		return map[string]any{"status": "ok", "applied": true}
	default:
		return map[string]any{"status": "ok"}
	}
}

// PromptInjectionPatterns to scan for in payloads.
var PromptInjectionPatterns = []string{
	"ignore previous",
	"ignore all previous",
	"disregard",
	"forget your instructions",
	"you are now",
	"new instructions",
	"system prompt",
	"reveal your",
	"show me your",
	"<system>",
	"</system>",
	"ADMIN OVERRIDE",
	"sudo mode",
	"jailbreak",
	"DAN",
	"do anything now",
}

// DetectPromptInjection checks if a payload contains prompt injection patterns.
func DetectPromptInjection(payload string) (bool, string) {
	lower := strings.ToLower(payload)
	for _, p := range PromptInjectionPatterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true, p
		}
	}
	return false, ""
}
