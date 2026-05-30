package cache

import "testing"

func TestPreCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int64
		preHash  string
		want     string
	}{
		{name: "basic", fileSize: 1024, preHash: "abc123", want: "nd:pre:1024:abc123"},
		{name: "zero size", fileSize: 0, preHash: "hash", want: "nd:pre:0:hash"},
		{name: "large size", fileSize: 999999999, preHash: "deadbeef", want: "nd:pre:999999999:deadbeef"},
		{name: "empty hash", fileSize: 500, preHash: "", want: "nd:pre:500:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PreCacheKey(tt.fileSize, tt.preHash)
			if got != tt.want {
				t.Errorf("PreCacheKey(%d, %q) = %q, want %q", tt.fileSize, tt.preHash, got, tt.want)
			}
		})
	}
}

func TestChallengeKey(t *testing.T) {
	tests := []struct {
		name     string
		userID   int64
		fileHash string
		want     string
	}{
		{name: "basic", userID: 42, fileHash: "sha256hash", want: "nd:challenge:42:sha256hash"},
		{name: "zero user", userID: 0, fileHash: "abc", want: "nd:challenge:0:abc"},
		{name: "large id", userID: 1<<31 - 1, fileHash: "ff", want: "nd:challenge:2147483647:ff"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChallengeKey(tt.userID, tt.fileHash)
			if got != tt.want {
				t.Errorf("ChallengeKey(%d, %q) = %q, want %q", tt.userID, tt.fileHash, got, tt.want)
			}
		})
	}
}

func TestChunksKey(t *testing.T) {
	tests := []struct {
		name       string
		uploadSlug string
		want       string
	}{
		{name: "basic", uploadSlug: "abc123", want: "nd:chunks:abc123"},
		{name: "uuid-like", uploadSlug: "550e8400-e29b-41d4-a716-446655440000", want: "nd:chunks:550e8400-e29b-41d4-a716-446655440000"},
		{name: "empty", uploadSlug: "", want: "nd:chunks:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChunksKey(tt.uploadSlug)
			if got != tt.want {
				t.Errorf("ChunksKey(%q) = %q, want %q", tt.uploadSlug, got, tt.want)
			}
		})
	}
}

func TestMergeLockKey(t *testing.T) {
	tests := []struct {
		name     string
		fileHash string
		want     string
	}{
		{name: "basic", fileHash: "abc123", want: "nd:lock:merge:abc123"},
		{name: "empty", fileHash: "", want: "nd:lock:merge:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeLockKey(tt.fileHash)
			if got != tt.want {
				t.Errorf("MergeLockKey(%q) = %q, want %q", tt.fileHash, got, tt.want)
			}
		})
	}
}

func TestRateLimitKey(t *testing.T) {
	tests := []struct {
		name   string
		ip     string
		action string
		want   string
	}{
		{name: "basic", ip: "192.168.1.1", action: "login", want: "nd:rate:192.168.1.1:login"},
		{name: "ipv6", ip: "::1", action: "api", want: "nd:rate:::1:api"},
		{name: "empty fields", ip: "", action: "", want: "nd:rate::"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RateLimitKey(tt.ip, tt.action)
			if got != tt.want {
				t.Errorf("RateLimitKey(%q, %q) = %q, want %q", tt.ip, tt.action, got, tt.want)
			}
		})
	}
}

func TestMediaProgressKey(t *testing.T) {
	tests := []struct {
		name          string
		transcodeSlug string
		want          string
	}{
		{name: "basic", transcodeSlug: "job123", want: "nd:media:transcode:job123:progress"},
		{name: "empty", transcodeSlug: "", want: "nd:media:transcode::progress"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MediaProgressKey(tt.transcodeSlug)
			if got != tt.want {
				t.Errorf("MediaProgressKey(%q) = %q, want %q", tt.transcodeSlug, got, tt.want)
			}
		})
	}
}
