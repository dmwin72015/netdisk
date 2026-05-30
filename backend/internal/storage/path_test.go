package storage

import (
	"testing"
)

func TestStoragePath(t *testing.T) {
	hash := "abcdef01234567890abcdef01234567890abcdef01234567890abcdef01234567890"
	tests := []struct {
		name     string
		hash     string
		filesDir string
		want     string
	}{
		{
			name:     "standard hash",
			hash:     hash,
			filesDir: "files",
			want:     "files/ab/cd/" + hash,
		},
		{
			name:     "different base dir",
			hash:     hash,
			filesDir: "storage/data",
			want:     "storage/data/ab/cd/" + hash,
		},
		{
			name:     "hash starting with zeros",
			hash:     "00abcdef01234567890abcdef01234567890abcdef01234567890abcdef01234",
			filesDir: "files",
			want:     "files/00/ab/" + "00abcdef01234567890abcdef01234567890abcdef01234567890abcdef01234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StoragePath(tt.hash, tt.filesDir)
			if got != tt.want {
				t.Errorf("StoragePath(%q, %q) = %q, want %q", tt.hash, tt.filesDir, got, tt.want)
			}
		})
	}
}

func TestAbsPath(t *testing.T) {
	hash := "abcdef01234567890abcdef01234567890abcdef01234567890abcdef01234567890"
	tests := []struct {
		name string
		root string
		hash string
		dir  string
		want string
	}{
		{
			name: "absolute root",
			root: "/data",
			hash: hash,
			dir:  "files",
			want: "/data/files/ab/cd/" + hash,
		},
		{
			name: "relative root",
			root: "var/uploads",
			hash: hash,
			dir:  "files",
			want: "var/uploads/files/ab/cd/" + hash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AbsPath(tt.root, tt.hash, tt.dir)
			if got != tt.want {
				t.Errorf("AbsPath(%q, %q, %q) = %q, want %q", tt.root, tt.hash, tt.dir, got, tt.want)
			}
		})
	}
}

func TestValidateHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
		want bool
	}{
		{name: "valid lowercase hex 64 chars", hash: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", want: true},
		{name: "valid all digits", hash: "0123456789012345678901234567890123456789012345678901234567890123", want: true},
		{name: "valid all a-f", hash: "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd", want: true},
		{name: "empty string", hash: "", want: false},
		{name: "too short", hash: "abcdef0123456789", want: false},
		{name: "too long (65 chars)", hash: "abcdef01234567890abcdef01234567890abcdef01234567890abcdef01234567890a", want: false},
		{name: "uppercase hex rejected", hash: "ABCDEF01234567890ABCDEF01234567890ABCDEF01234567890ABCDEF01234567890", want: false},
		{name: "contains non-hex char", hash: "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg", want: false},
		{name: "contains spaces", hash: "abcdef0123456789  abcdef0123456789  abcdef0123456789  abcdef0123456789", want: false},
		{name: "contains dashes", hash: "abcdef01-23456789-0abcdef0123456789-0abcdef01234567890abcdef01234567", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateHash(tt.hash)
			if got != tt.want {
				t.Errorf("ValidateHash(%q) = %v, want %v", tt.hash, got, tt.want)
			}
		})
	}
}

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		name string
		slug string
		want bool
	}{
		{name: "valid alphanumeric slug", slug: "abcdefghijklmnopqrstu", want: true},
		{name: "valid with digits", slug: "a1b2c3d4e5f6g7h8i9j0k", want: true},
		{name: "valid with underscores", slug: "a_b_c_d_e_f_g_h_i_j_k", want: true},
		{name: "valid with dashes", slug: "a-b-c-d-e-f-g-h-i-j-k", want: true},
		{name: "valid mixed", slug: "Ab1_Cd2-Ef3_Gh4-Ij5_K", want: true},
		{name: "empty string", slug: "", want: false},
		{name: "too short (20 chars)", slug: "abcdefghijklmnopqrst", want: false},
		{name: "too long (22 chars)", slug: "abcdefghijklmnopqrstuV", want: false},
		{name: "contains dot", slug: "abcdefghijklmnopqrs.u", want: false},
		{name: "contains space", slug: "abcdefghijklmnopqrs u", want: false},
		{name: "contains slash", slug: "abcdefghijklmnopqrs/u", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateSlug(tt.slug)
			if got != tt.want {
				t.Errorf("ValidateSlug(%q) = %v, want %v", tt.slug, got, tt.want)
			}
		})
	}
}

func TestSafePath(t *testing.T) {
	tests := []struct {
		name    string
		subPath string
		want    string
	}{
		{name: "simple filename", subPath: "document.pdf", want: "document.pdf"},
		{name: "nested path", subPath: "folder/subfolder/file.txt", want: "folder/subfolder/file.txt"},
		{name: "dot segments cleaned", subPath: "folder/../file.txt", want: "file.txt"},
		{name: "path traversal with dot-dot", subPath: "../../../etc/passwd", want: ""},
		{name: "absolute path rejected", subPath: "/etc/passwd", want: ""},
		{name: "single dot returns empty", subPath: ".", want: ""},
		{name: "double dot returns empty", subPath: "..", want: ""},
		{name: "dot-dot with path", subPath: "../file.txt", want: ""},
		{name: "embedded traversal", subPath: "folder/../../../etc/passwd", want: ""},
		{name: "trailing dot-dot", subPath: "folder/..", want: ""},
		{name: "empty string", subPath: "", want: ""},
		{name: "clean redundant slashes", subPath: "folder//file.txt", want: "folder/file.txt"},
		{name: "trailing slash removed", subPath: "folder/file.txt/", want: "folder/file.txt"},
		{name: "special characters allowed", subPath: "my file (copy) #1.txt", want: "my file (copy) #1.txt"},
		{name: "unicode allowed", subPath: "文件.txt", want: "文件.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafePath(tt.subPath)
			if got != tt.want {
				t.Errorf("SafePath(%q) = %q, want %q", tt.subPath, got, tt.want)
			}
		})
	}
}
