package geoip

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

// Info holds GeoIP data for an IP address.
type Info struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
	Org     string `json:"org"`
	Loc     string `json:"loc"`
}

type cacheEntry struct {
	info      Info
	expiresAt time.Time
}

// Client queries ipinfo.io for GeoIP data with an in-memory cache.
type Client struct {
	token      string
	httpClient *http.Client
	cache      sync.Map
	ttl        time.Duration
}

// NewClient creates a GeoIP client. Token is optional (empty = free tier).
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		ttl: 1 * time.Hour,
	}
}

// Lookup returns GeoIP info for the given IP. Returns zero Info on error (non-blocking).
func (c *Client) Lookup(ip string) Info {
	// Skip private/loopback IPs.
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified() {
		return Info{}
	}

	// Check cache.
	if entry, ok := c.cache.Load(ip); ok {
		ce := entry.(cacheEntry)
		if time.Now().Before(ce.expiresAt) {
			return ce.info
		}
		c.cache.Delete(ip)
	}

	// Query ipinfo.io.
	url := fmt.Sprintf("https://ipinfo.io/%s/json", ip)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Info{}
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Debug("geoip lookup failed", "ip", ip, "error", err)
		return Info{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug("geoip lookup non-200", "ip", ip, "status", resp.StatusCode)
		return Info{}
	}

	var info Info
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return Info{}
	}

	// Cache result.
	c.cache.Store(ip, cacheEntry{
		info:      info,
		expiresAt: time.Now().Add(c.ttl),
	})

	return info
}

// Attrs returns slog attributes for the geo info (only non-empty fields).
func (info Info) Attrs() []slog.Attr {
	var attrs []slog.Attr
	if info.Country != "" {
		attrs = append(attrs, slog.String("geo_country", info.Country))
	}
	if info.City != "" {
		attrs = append(attrs, slog.String("geo_city", info.City))
	}
	if info.Region != "" {
		attrs = append(attrs, slog.String("geo_region", info.Region))
	}
	if info.Org != "" {
		attrs = append(attrs, slog.String("geo_org", info.Org))
	}
	if info.Loc != "" {
		attrs = append(attrs, slog.String("geo_loc", info.Loc))
	}
	return attrs
}
