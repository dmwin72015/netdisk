package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

func TestHashToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{name: "basic string", token: "my-refresh-token-123"},
		{name: "empty string", token: ""},
		{name: "long token", token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"},
		{name: "special chars", token: "a!@#$%^&*()_+-=[]{}|;':\",./<>?`~"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashToken(tt.token)

			// Verify it matches manual SHA-256 computation
			h := sha256.Sum256([]byte(tt.token))
			want := hex.EncodeToString(h[:])
			if got != want {
				t.Errorf("hashToken(%q) = %q, want %q", tt.token, got, want)
			}

			// Verify it's a valid 64-char hex string (SHA-256 = 32 bytes = 64 hex chars)
			if len(got) != 64 {
				t.Errorf("hashToken output length = %d, want 64", len(got))
			}
		})
	}
}

func TestHashTokenDeterministic(t *testing.T) {
	token := "some-token"
	a := hashToken(token)
	b := hashToken(token)
	if a != b {
		t.Errorf("hashToken not deterministic: %q != %q", a, b)
	}
}

func TestHashTokenDifferentInputs(t *testing.T) {
	a := hashToken("token-a")
	b := hashToken("token-b")
	if a == b {
		t.Error("different inputs should produce different hashes")
	}
}

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "duplicate key", err: errors.New("ERROR: duplicate key value violates unique constraint"), want: true},
		{name: "sqlstate 23505", err: errors.New("pq: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)"), want: true},
		{name: "regular error", err: errors.New("something went wrong"), want: false},
		{name: "empty error message", err: errors.New(""), want: false},
		{name: "23505 in middle", err: errors.New("pgx: error 23505: duplicate"), want: true},
		{name: "duplicate key substring", err: errors.New("duplicate key violation"), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUniqueViolation(tt.err)
			if got != tt.want {
				t.Errorf("isUniqueViolation(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{name: "found at start", s: "hello world", substr: "hello", want: true},
		{name: "found at end", s: "hello world", substr: "world", want: true},
		{name: "found in middle", s: "hello world", substr: "lo wo", want: true},
		{name: "not found", s: "hello world", substr: "xyz", want: false},
		{name: "exact match", s: "abc", substr: "abc", want: true},
		{name: "empty substr", s: "abc", substr: "", want: true},
		{name: "both empty", s: "", substr: "", want: true},
		{name: "substr longer than s", s: "ab", substr: "abc", want: false},
		{name: "empty s non-empty substr", s: "", substr: "a", want: false},
		{name: "single char match", s: "abc", substr: "b", want: true},
		{name: "single char no match", s: "abc", substr: "d", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestContainsSubstr(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{name: "match at start", s: "abcdef", substr: "abc", want: true},
		{name: "match at end", s: "abcdef", substr: "def", want: true},
		{name: "no match", s: "abcdef", substr: "xyz", want: false},
		{name: "single char", s: "abcdef", substr: "c", want: true},
		{name: "repeated pattern", s: "ababab", substr: "bab", want: true},
		{name: "substr equals s", s: "test", substr: "test", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsSubstr(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("containsSubstr(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}
