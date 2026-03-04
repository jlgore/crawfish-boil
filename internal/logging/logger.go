package logging

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
)

// Init configures the global slog logger for structured JSON output.
func Init() {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(h))
}

// Event logs a honeypot event with standard fields.
func Event(ctx context.Context, event string, attrs ...slog.Attr) {
	args := make([]any, 0, len(attrs)*2+2)
	args = append(args, slog.String("event", event))
	for _, a := range attrs {
		args = append(args, a)
	}
	slog.InfoContext(ctx, event, args...)
}

// SourceIP extracts the real client IP from an HTTP request.
// Checks Cf-Connecting-Ip (Cloudflare), X-Real-Ip, X-Forwarded-For, then RemoteAddr.
func SourceIP(r *http.Request) (string, string) {
	// Cloudflare sets this to the true client IP.
	if ip := r.Header.Get("Cf-Connecting-Ip"); ip != "" {
		return ip, ""
	}
	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		return ip, ""
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// First IP in the chain is the client.
		if i := strings.Index(xff, ","); i != -1 {
			return strings.TrimSpace(xff[:i]), ""
		}
		return strings.TrimSpace(xff), ""
	}
	host, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr, ""
	}
	return host, port
}
