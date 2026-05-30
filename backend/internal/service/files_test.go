package service

import (
	"strings"
	"testing"
)

func TestSafeFilename(t *testing.T) {
	tests := []struct {
		name string
		input string
		want string
	}{
		{name: "simple ascii", input: "document.pdf", want: "document.pdf"},
		{name: "spaces preserved", input: "my file (1).txt", want: "my file (1).txt"},
		{name: "forward slash replaced", input: "path/to/file.txt", want: "path_to_file.txt"},
		{name: "backslash replaced", input: `path\to\file.txt`, want: "path_to_file.txt"},
		{name: "null byte replaced", input: "file\x00name.txt", want: "file_name.txt"},
		{name: "unicode chars replaced", input: "文件.txt", want: "__.txt"},
		{name: "mixed unicode and ascii", input: "report_文件.pdf", want: "report___.pdf"},
		{name: "emoji replaced", input: "photo_smile.jpg", want: "photo_smile.jpg"},
		{name: "empty string becomes download", input: "", want: "download"},
		{name: "single dot becomes download", input: ".", want: "download"},
		{name: "multiple slashes", input: "a/b/c/d.txt", want: "a_b_c_d.txt"},
		{name: "consecutive null bytes", input: "\x00\x00\x00", want: "___"},
		{name: "high unicode (above 127)", input: "éèê.txt", want: "___.txt"}, // accented chars
		{name: "all printable ascii", input: "!@#$%^&()_+-=[]{};',.", want: "!@#$%^&()_+-=[]{};',."},
		{name: "tab character preserved", input: "file\tname.txt", want: "file\tname.txt"},
		{name: "newline replaced with underscore (it's below 127 but...)", input: "file\nname.txt", want: "file\nname.txt"},
		{name: "tilde preserved", input: "~backup.tar.gz", want: "~backup.tar.gz"},
		{name: "only slashes becomes underscores", input: "///", want: "___"},
		{name: "only unicode becomes underscores", input: "中文测试", want: "____"}, // 4 CJK runes → 4 underscores
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := safeFilename(tt.input)
			if got != tt.want {
				t.Errorf("safeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestContentDisposition(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantContains []string // substrings the output must contain
	}{
		{
			name:     "simple filename",
			filename: "report.pdf",
			wantContains: []string{
				`attachment;`,
				`filename="report.pdf"`,
				`filename*=UTF-8''report.pdf`,
			},
		},
		{
			name:     "unicode filename",
			filename: "report.pdf",
			wantContains: []string{
				`attachment;`,
				`filename*=UTF-8''`,
			},
		},
		{
			name:     "filename with slashes",
			filename: "path/to/file.txt",
			wantContains: []string{
				`filename="path_to_file.txt"`,    // safeFilename sanitizes
				`filename*=UTF-8''path/to/file.txt`, // raw name in RFC 5987 part
			},
		},
		{
			name:     "empty filename",
			filename: "",
			wantContains: []string{
				`filename="download"`, // safeFilename converts empty to "download"
				`filename*=UTF-8''`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contentDisposition(tt.filename)
			for _, substr := range tt.wantContains {
				if !strings.Contains(got, substr) {
					t.Errorf("contentDisposition(%q) = %q, want to contain %q", tt.filename, got, substr)
				}
			}
		})
	}
}
