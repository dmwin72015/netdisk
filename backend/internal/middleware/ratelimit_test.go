package middleware

import "testing"

func TestIsLoopback(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{name: "empty string", ip: "", want: true},
		{name: "ipv4 loopback 127.0.0.1", ip: "127.0.0.1", want: true},
		{name: "ipv4 loopback 127.0.0.2", ip: "127.0.0.2", want: true},
		{name: "ipv4 loopback 127.255.255.255", ip: "127.255.255.255", want: true},
		{name: "private 10.x", ip: "10.0.0.1", want: false},
		{name: "private 192.168.x", ip: "192.168.1.1", want: false},
		{name: "public ip", ip: "8.8.8.8", want: false},
		{name: "ipv6 loopback", ip: "::1", want: true},
		{name: "ipv4-mapped loopback", ip: "::ffff:127.0.0.1", want: true},
		{name: "invalid ip", ip: "not-an-ip", want: false},
		{name: "ipv6 private", ip: "fe80::1", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLoopback(tt.ip)
			if got != tt.want {
				t.Errorf("isLoopback(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
