package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRotatingFileWriterRotatesBySize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "server.log")
	writer, err := newRotatingFileWriter(path, 10)
	if err != nil {
		t.Fatalf("newRotatingFileWriter() error = %v", err)
	}
	defer writer.Close()

	if _, err := writer.Write(bytes.Repeat([]byte("a"), 8)); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	if _, err := writer.Write(bytes.Repeat([]byte("b"), 8)); err != nil {
		t.Fatalf("second Write() error = %v", err)
	}

	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read current log: %v", err)
	}
	if string(current) != "bbbbbbbb" {
		t.Fatalf("current log = %q, want %q", current, "bbbbbbbb")
	}

	matches, err := filepath.Glob(filepath.Join(dir, "server-*.log"))
	if err != nil {
		t.Fatalf("glob rotated logs: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("rotated log count = %d, want 1", len(matches))
	}
	rotated, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read rotated log: %v", err)
	}
	if string(rotated) != "aaaaaaaa" {
		t.Fatalf("rotated log = %q, want %q", rotated, "aaaaaaaa")
	}
}
