package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateChunksReportsMissingAndSizeIssues(t *testing.T) {
	root := t.TempDir()
	store := NewLocal(root, "tmp", "files")
	uploadSlug := "upload-12345678901234"
	chunkDir := filepath.Join(root, "tmp", uploadSlug)
	if err := os.MkdirAll(chunkDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(chunkDir, "chunk_000000"), []byte("1234"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(chunkDir, "chunk_000002"), []byte("12"), 0o644); err != nil {
		t.Fatal(err)
	}

	issues, err := store.ValidateChunks(uploadSlug, 3, 4, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d: %#v", len(issues), issues)
	}
	if issues[0].Index != 1 || !issues[0].Missing || issues[0].Expected != 4 {
		t.Fatalf("unexpected issue: %#v", issues[0])
	}

	valid, err := store.ValidChunkSet(uploadSlug, 3, 4, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !valid[0] || !valid[2] || valid[1] {
		t.Fatalf("unexpected valid chunk set: %#v", valid)
	}
}

func TestValidateChunksReportsLastChunkSizeIssue(t *testing.T) {
	root := t.TempDir()
	store := NewLocal(root, "tmp", "files")
	uploadSlug := "upload-12345678901234"
	chunkDir := filepath.Join(root, "tmp", uploadSlug)
	if err := os.MkdirAll(chunkDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(chunkDir, "chunk_000000"), []byte("1234"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(chunkDir, "chunk_000001"), []byte("1234"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(chunkDir, "chunk_000002"), []byte("123"), 0o644); err != nil {
		t.Fatal(err)
	}

	issues, err := store.ValidateChunks(uploadSlug, 3, 4, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d: %#v", len(issues), issues)
	}
	if issues[0].Index != 2 || issues[0].Missing || issues[0].Expected != 2 || issues[0].Actual != 3 {
		t.Fatalf("unexpected issue: %#v", issues[0])
	}
}
