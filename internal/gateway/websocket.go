package gateway

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"nhooyr.io/websocket"

	"openclaw-honeypot/internal/logging"
	"openclaw-honeypot/internal/simulation"
)

const (
	protocolVersion = 3
	tickInterval    = 15 * time.Second
)

type wsEvent struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
	ID      string         `json:"id,omitempty"`
	Error   *wsError       `json:"error,omitempty"`
}

type wsError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	srcIP, srcPort := logging.SourceIP(r)
	sess := simulation.NewSession()
	defer simulation.RemoveSession(sess.ID)

	logging.Event(r.Context(), "ws_connect",
		slog.String("src_ip", srcIP),
		slog.String("src_port", srcPort),
		slog.String("session_id", sess.ID),
		slog.String("user_agent", r.UserAgent()),
	)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Accept any origin.
	})
	if err != nil {
		slog.Error("websocket accept failed", "error", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Send challenge.
	nonce := generateNonce()
	challenge := wsEvent{
		Type: "connect.challenge",
		Payload: map[string]any{
			"nonce":     nonce,
			"timestamp": time.Now().UnixMilli(),
			"protocol":  protocolVersion,
			"server":    "openclaw/0.14.2",
		},
	}
	if err := writeJSON(ctx, conn, challenge); err != nil {
		return
	}
	logging.Event(ctx, "ws_challenge_sent",
		slog.String("src_ip", srcIP),
		slog.String("session_id", sess.ID),
		slog.String("nonce", nonce),
	)

	// Wait for connect message.
	conn.SetReadLimit(1 << 20) // 1MB
	_, msg, err := conn.Read(ctx)
	if err != nil {
		logging.Event(ctx, "ws_disconnect",
			slog.String("src_ip", srcIP),
			slog.String("session_id", sess.ID),
			slog.String("reason", "read_error_during_auth"),
		)
		return
	}

	// Parse connect request.
	var connectReq struct {
		Type    string `json:"type"`
		Payload struct {
			Auth struct {
				Mode  string `json:"mode"`
				Token string `json:"token"`
				Pass  string `json:"password"`
				User  string `json:"username"`
			} `json:"auth"`
			Client struct {
				ID       string `json:"id"`
				Version  string `json:"version"`
				Platform string `json:"platform"`
			} `json:"client"`
			Nonce string `json:"nonce"`
		} `json:"payload"`
	}

	rawPayload := string(msg)
	json.Unmarshal(msg, &connectReq)

	// Determine auth value.
	authMode := connectReq.Payload.Auth.Mode
	if authMode == "" {
		authMode = "unknown"
	}
	authValue := connectReq.Payload.Auth.Token
	if authValue == "" {
		authValue = connectReq.Payload.Auth.Pass
	}
	if authValue == "" {
		authValue = connectReq.Payload.Auth.User
	}

	sess.AuthMode = authMode
	sess.AuthValue = authValue
	sess.ClientInfo = map[string]any{
		"id":       connectReq.Payload.Client.ID,
		"version":  connectReq.Payload.Client.Version,
		"platform": connectReq.Payload.Client.Platform,
	}

	logging.Event(ctx, "ws_auth_attempt",
		slog.String("src_ip", srcIP),
		slog.String("src_port", srcPort),
		slog.String("session_id", sess.ID),
		slog.String("auth_mode", authMode),
		slog.String("auth_value", authValue),
		slog.Int("protocol_version", protocolVersion),
		slog.Any("client_info", sess.ClientInfo),
		slog.String("user_agent", r.UserAgent()),
		slog.String("raw_payload", rawPayload),
	)

	// Check for prompt injection in auth payload.
	if detected, pattern := simulation.DetectPromptInjection(rawPayload); detected {
		logging.Event(ctx, "prompt_injection",
			slog.String("src_ip", srcIP),
			slog.String("session_id", sess.ID),
			slog.String("pattern", pattern),
			slog.String("raw_payload", rawPayload),
		)
	}

	// Simulate auth delay then accept.
	time.Sleep(simulation.AuthDelay())
	sess.Authenticated = true

	logging.Event(ctx, "ws_auth_success",
		slog.String("src_ip", srcIP),
		slog.String("session_id", sess.ID),
		slog.String("auth_mode", authMode),
	)

	// Send hello-ok.
	helloOk := wsEvent{
		Type: "hello-ok",
		Payload: map[string]any{
			"protocol":       protocolVersion,
			"tickIntervalMs": int(tickInterval.Milliseconds()),
			"session":        sess.ID,
			"deviceToken":    "dt_" + generateNonce()[:16],
			"server": map[string]any{
				"version": "0.14.2",
				"runtime": "node/22.11.0",
			},
		},
	}
	if err := writeJSON(ctx, conn, helloOk); err != nil {
		return
	}

	// Start tick goroutine.
	go func() {
		ticker := time.NewTicker(tickInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				tick := wsEvent{
					Type: "tick",
					Payload: map[string]any{
						"ts":      t.UnixMilli(),
						"clients": 1,
					},
				}
				if writeJSON(ctx, conn, tick) != nil {
					return
				}
			}
		}
	}()

	// RPC loop.
	connectTime := time.Now()
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			duration := time.Since(connectTime)
			closeCode := websocket.CloseStatus(err)
			logging.Event(ctx, "ws_disconnect",
				slog.String("src_ip", srcIP),
				slog.String("session_id", sess.ID),
				slog.Int("close_code", int(closeCode)),
				slog.Float64("duration_s", duration.Seconds()),
			)
			return
		}

		var rpc struct {
			Type    string          `json:"type"`
			ID      string          `json:"id"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
			Payload json.RawMessage `json:"payload"`
		}
		json.Unmarshal(msg, &rpc)

		rawRPC := string(msg)
		method := rpc.Method
		if method == "" {
			method = rpc.Type
		}

		logging.Event(ctx, "ws_rpc_call",
			slog.String("src_ip", srcIP),
			slog.String("session_id", sess.ID),
			slog.String("method", method),
			slog.String("raw_payload", rawRPC),
		)

		// Log full payload for interesting methods.
		if isInterestingMethod(method) {
			logging.Event(ctx, "ws_rpc_payload",
				slog.String("src_ip", srcIP),
				slog.String("session_id", sess.ID),
				slog.String("method", method),
				slog.String("raw_payload", rawRPC),
			)
		}

		// Check for prompt injection.
		if detected, pattern := simulation.DetectPromptInjection(rawRPC); detected {
			logging.Event(ctx, "prompt_injection",
				slog.String("src_ip", srcIP),
				slog.String("session_id", sess.ID),
				slog.String("method", method),
				slog.String("pattern", pattern),
				slog.String("raw_payload", rawRPC),
			)
		}

		// Send response.
		params := rpc.Params
		if params == nil {
			params = rpc.Payload
		}
		result := simulation.RPCResponse(method, params)
		resp := wsEvent{
			Type:    "response",
			ID:      rpc.ID,
			Payload: result,
		}
		if err := writeJSON(ctx, conn, resp); err != nil {
			return
		}
	}
}

func writeJSON(ctx context.Context, conn *websocket.Conn, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, b)
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func isInterestingMethod(method string) bool {
	interesting := []string{
		"tools.invoke", "chat.send", "chat.message",
		"config.patch", "config.set", "shell.exec",
		"code.edit", "file.write",
	}
	m := strings.ToLower(method)
	for _, im := range interesting {
		if strings.Contains(m, im) {
			return true
		}
	}
	return false
}

// FormatCloseCode returns descriptive close code for logging (unused but available).
func FormatCloseCode(code int) string {
	switch code {
	case 4008:
		return "protocol_disconnect"
	case 1008:
		return "auth_failure"
	default:
		return fmt.Sprintf("code_%d", code)
	}
}
