package useragent

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		wantOS    string
		wantBrow  string
	}{
		{
			name:      "Chrome on Windows",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			wantOS:    "Windows 10",
			wantBrow:  "Chrome",
		},
		{
			name:      "Safari on macOS",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
			wantOS:    "Intel Mac OS X 14_0",
			wantBrow:  "Safari",
		},
		{
			name:      "Firefox on Linux",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0",
			wantOS:    "Linux x86_64",
			wantBrow:  "Firefox",
		},
		{
			name:      "Chrome on Android",
			userAgent: "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
			wantOS:    "Android 14",
			wantBrow:  "Chrome",
		},
		{
			name:      "empty string",
			userAgent: "",
			wantOS:    "",
			wantBrow:  "",
		},
		{
			name:      "bot user agent",
			userAgent: "Googlebot/2.1 (+http://www.google.com/bot.html)",
			wantOS:    "",
			wantBrow:  "Googlebot",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := Parse(tt.userAgent)
			if info.OS != tt.wantOS {
				t.Errorf("Parse(%q).OS = %q, want %q", tt.userAgent, info.OS, tt.wantOS)
			}
			if info.Browser != tt.wantBrow {
				t.Errorf("Parse(%q).Browser = %q, want %q", tt.userAgent, info.Browser, tt.wantBrow)
			}
		})
	}
}
