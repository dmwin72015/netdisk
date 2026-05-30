package storage

import (
	"path/filepath"
	"regexp"
	"strings"
)

var hashRe = regexp.MustCompile(`^[a-f0-9]{64}$`)
var slugRe = regexp.MustCompile(`^[A-Za-z0-9_-]{21}$`)

// StoragePath returns the two-level directory path for a SHA-256 hash.
// e.g. ("abcdef...", "files") -> "files/ab/cd/abcdef..."
func StoragePath(fileHash, filesDir string) string {
	return filepath.Join(filesDir, fileHash[0:2], fileHash[2:4], fileHash)
}

// AbsPath returns the full absolute path for a file hash under the given root.
func AbsPath(root, fileHash, filesDir string) string {
	return filepath.Join(root, StoragePath(fileHash, filesDir))
}

// ValidateHash checks that a file hash is exactly 64 lowercase hex characters.
func ValidateHash(fileHash string) bool {
	return hashRe.MatchString(fileHash)
}

// ValidateSlug checks that a slug is exactly 21 URL-safe characters.
func ValidateSlug(slug string) bool {
	return slugRe.MatchString(slug)
}

// SafePath cleans a sub-path and ensures it doesn't escape the parent directory.
func SafePath(subPath string) string {
	cleaned := filepath.Clean(subPath)
	if cleaned == "." || cleaned == ".." {
		return ""
	}
	if strings.HasPrefix(cleaned, "..") || strings.HasPrefix(cleaned, "/") {
		return ""
	}
	return cleaned
}
