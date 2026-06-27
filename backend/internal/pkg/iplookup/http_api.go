package iplookup

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// HTTPLookup queries a third-party HTTP API for IP geolocation.
type HTTPLookup struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

// NewHTTPLookup creates a new HTTP-based IP lookup client.
// endpoint should contain a {ip} placeholder, e.g. "http://ip-api.com/json/{ip}".
func NewHTTPLookup(endpoint, apiKey string) *HTTPLookup {
	return &HTTPLookup{
		endpoint: endpoint,
		apiKey:   apiKey,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (h *HTTPLookup) Lookup(ip string) (Region, error) {
	// Skip private/loopback IPs
	if parsed := net.ParseIP(ip); parsed != nil && (parsed.IsLoopback() || parsed.IsPrivate()) {
		return Region{}, nil
	}

	url := strings.Replace(h.endpoint, "{ip}", ip, 1)
	if h.apiKey != "" {
		if strings.Contains(url, "?") {
			url += "&key=" + h.apiKey
		} else {
			url += "?key=" + h.apiKey
		}
	}

	resp, err := h.client.Get(url)
	if err != nil {
		return Region{}, fmt.Errorf("ip lookup request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Region{}, fmt.Errorf("ip lookup status: %d", resp.StatusCode)
	}

	var result struct {
		Country  string `json:"country"`
		Region   string `json:"region"`
		RegionName string `json:"regionName"`
		City     string `json:"city"`
		ISP      string `json:"isp"`
		Query    string `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Region{}, fmt.Errorf("ip lookup decode: %w", err)
	}

	return Region{
		Country:  result.Country,
		Province: result.RegionName,
		City:     result.City,
		ISP:      result.ISP,
	}, nil
}
