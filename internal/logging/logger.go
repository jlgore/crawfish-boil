package logging

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
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

// SourceIP extracts the remote IP and port from an HTTP request.
func SourceIP(r *http.Request) (string, string) {
	host, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr, ""
	}
	return host, port
}
