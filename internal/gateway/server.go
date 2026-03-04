package gateway

import (
	"log/slog"
	"net/http"
	"strings"

	"openclaw-honeypot/internal/geoip"
)

// Server is the OpenClaw honeypot gateway.
type Server struct {
	Addr string
	Geo  *geoip.Client
}

// NewServer creates a gateway server on the given address.
func NewServer(addr string, geo *geoip.Client) *Server {
	return &Server{Addr: addr, Geo: geo}
}

// ListenAndServe starts the HTTP + WebSocket server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.route)

	slog.Info("openclaw honeypot starting",
		"addr", s.Addr,
		"protocol", protocolVersion,
		"version", "0.14.2",
	)
	return http.ListenAndServe(s.Addr, mux)
}

func (s *Server) route(w http.ResponseWriter, r *http.Request) {
	// Add common headers to look realistic.
	w.Header().Set("Server", "OpenClaw/0.14.2")
	w.Header().Set("X-Request-Id", generateNonce()[:12])

	// WebSocket upgrade detection.
	if isWebSocketUpgrade(r) {
		s.handleWebSocket(w, r)
		return
	}

	s.handleHTTP(w, r)
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}
