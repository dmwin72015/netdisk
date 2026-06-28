package iplookup

import "strings"

// Region represents an IP geolocation result.
type Region struct {
	Country  string
	Province string
	City     string
	ISP      string
}

// String returns a pipe-separated region string, omitting empty segments.
func (r Region) String() string {
	parts := make([]string, 0, 4)
	if r.Country != "" {
		parts = append(parts, r.Country)
	}
	if r.Province != "" {
		parts = append(parts, r.Province)
	}
	if r.City != "" {
		parts = append(parts, r.City)
	}
	if r.ISP != "" {
		parts = append(parts, r.ISP)
	}
	return strings.Join(parts, "|")
}

// Lookup defines the interface for IP geolocation queries.
type Lookup interface {
	Lookup(ip string) (Region, error)
}

// noopLookup is a fallback that always returns empty region.
type noopLookup struct{}

func NewNoopLookup() Lookup { return noopLookup{} }

func (noopLookup) Lookup(_ string) (Region, error) { return Region{}, nil }
