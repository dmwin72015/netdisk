package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type ChunkIssue struct {
	Index    int
	Expected int64
	Actual   int64
	Missing  bool
}

// Local implements the Storage interface using the local filesystem.
type Local struct {
	root     string
	tmpDir   string
	filesDir string
}

// NewLocal creates a new Local storage with the given root directory.
func NewLocal(root, tmpDir, filesDir string) *Local {
	return &Local{root: root, tmpDir: tmpDir, filesDir: filesDir}
}

// WriteChunk writes a single chunk to the temporary upload directory.
func (s *Local) WriteChunk(uploadSlug string, chunkIndex int, data io.Reader) error {
	if !ValidateSlug(uploadSlug) {
		return fmt.Errorf("invalid upload slug")
	}
	dir := filepath.Join(s.root, s.tmpDir, uploadSlug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir chunk dir: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("chunk_%06d", chunkIndex))
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create chunk: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, data); err != nil {
		return fmt.Errorf("write chunk: %w", err)
	}
	return nil
}

// ListChunks returns the list of chunk indices that have been uploaded.
func (s *Local) ListChunks(uploadSlug string) ([]int, error) {
	if !ValidateSlug(uploadSlug) {
		return nil, fmt.Errorf("invalid upload slug")
	}
	dir := filepath.Join(s.root, s.tmpDir, uploadSlug)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read chunk dir: %w", err)
	}
	var indices []int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "chunk_") {
			continue
		}
		idx, err := strconv.Atoi(name[6:])
		if err != nil {
			continue
		}
		indices = append(indices, idx)
	}
	sort.Ints(indices)
	return indices, nil
}

func (s *Local) ValidateChunks(uploadSlug string, totalChunks int, chunkSize int64, fileSize int64) ([]ChunkIssue, error) {
	_, issues, err := s.inspectChunks(uploadSlug, totalChunks, chunkSize, fileSize)
	return issues, err
}

func (s *Local) ValidChunkSet(uploadSlug string, totalChunks int, chunkSize int64, fileSize int64) (map[int]bool, error) {
	valid, _, err := s.inspectChunks(uploadSlug, totalChunks, chunkSize, fileSize)
	return valid, err
}

func (s *Local) inspectChunks(uploadSlug string, totalChunks int, chunkSize int64, fileSize int64) (map[int]bool, []ChunkIssue, error) {
	if !ValidateSlug(uploadSlug) {
		return nil, nil, fmt.Errorf("invalid upload slug")
	}
	if totalChunks <= 0 || chunkSize <= 0 || fileSize <= 0 {
		return nil, nil, fmt.Errorf("invalid chunk metadata")
	}

	chunkDir := filepath.Join(s.root, s.tmpDir, uploadSlug)
	valid := make(map[int]bool)
	issues := make([]ChunkIssue, 0)
	for i := 0; i < totalChunks; i++ {
		expected := chunkSize
		if i == totalChunks-1 {
			expected = fileSize - int64(i)*chunkSize
		}
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%06d", i))
		info, err := os.Stat(chunkPath)
		if err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, ChunkIssue{Index: i, Expected: expected, Missing: true})
				continue
			}
			return nil, nil, fmt.Errorf("stat chunk %d: %w", i, err)
		}
		if info.Size() == expected {
			valid[i] = true
		} else {
			issues = append(issues, ChunkIssue{Index: i, Expected: expected, Actual: info.Size()})
		}
	}
	return valid, issues, nil
}

// MergeChunks combines all chunks into a single file, computes SHA-256,
// verifies against the expected hash, and atomically renames to final path.
// Returns the final storage path on success.
func (s *Local) MergeChunks(uploadSlug string, fileHash string, totalChunks int) (string, error) {
	if !ValidateSlug(uploadSlug) {
		return "", fmt.Errorf("invalid upload slug")
	}
	if !ValidateHash(fileHash) {
		return "", fmt.Errorf("invalid file hash")
	}

	finalDir := filepath.Join(s.root, StoragePath(fileHash, s.filesDir))
	if err := os.MkdirAll(filepath.Dir(finalDir), 0o755); err != nil {
		return "", fmt.Errorf("mkdir final dir: %w", err)
	}

	// Write to a temporary .merging file first.
	mergingPath := finalDir + ".merging"
	out, err := os.Create(mergingPath)
	if err != nil {
		return "", fmt.Errorf("create merging file: %w", err)
	}

	h := sha256.New()
	writer := io.MultiWriter(out, h)

	chunkDir := filepath.Join(s.root, s.tmpDir, uploadSlug)
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%06d", i))
		f, err := os.Open(chunkPath)
		if err != nil {
			out.Close()
			os.Remove(mergingPath)
			return "", fmt.Errorf("open chunk %d: %w", i, err)
		}
		if _, err := io.Copy(writer, f); err != nil {
			f.Close()
			out.Close()
			os.Remove(mergingPath)
			return "", fmt.Errorf("copy chunk %d: %w", i, err)
		}
		f.Close()
	}
	out.Close()

	// Verify SHA-256.
	actual := hex.EncodeToString(h.Sum(nil))
	if actual != fileHash {
		os.Remove(mergingPath)
		return "", fmt.Errorf("hash mismatch: expected %s, got %s", fileHash, actual)
	}

	// Atomic rename to final path.
	if err := os.Rename(mergingPath, finalDir); err != nil {
		os.Remove(mergingPath)
		return "", fmt.Errorf("rename to final: %w", err)
	}

	return StoragePath(fileHash, s.filesDir), nil
}

// MergeChunksAndHash merges chunks, computes SHA-256, and returns the hash.
// Used when file hash is not known at init time (pre_hash fast path).
func (s *Local) MergeChunksAndHash(uploadSlug string, totalChunks int) (string, string, error) {
	if !ValidateSlug(uploadSlug) {
		return "", "", fmt.Errorf("invalid upload slug")
	}

	chunkDir := filepath.Join(s.root, s.tmpDir, uploadSlug)

	// Merge to temp file and compute hash.
	tmpPath := filepath.Join(s.root, s.tmpDir, uploadSlug+".merged")
	out, err := os.Create(tmpPath)
	if err != nil {
		return "", "", fmt.Errorf("create temp file: %w", err)
	}

	h := sha256.New()
	writer := io.MultiWriter(out, h)

	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%06d", i))
		f, err := os.Open(chunkPath)
		if err != nil {
			out.Close()
			os.Remove(tmpPath)
			return "", "", fmt.Errorf("open chunk %d: %w", i, err)
		}
		if _, err := io.Copy(writer, f); err != nil {
			f.Close()
			out.Close()
			os.Remove(tmpPath)
			return "", "", fmt.Errorf("copy chunk %d: %w", i, err)
		}
		f.Close()
	}
	out.Close()

	actual := hex.EncodeToString(h.Sum(nil))

	// Move to final path based on computed hash.
	finalDir := filepath.Join(s.root, StoragePath(actual, s.filesDir))
	if err := os.MkdirAll(filepath.Dir(finalDir), 0o755); err != nil {
		os.Remove(tmpPath)
		return "", "", fmt.Errorf("mkdir final dir: %w", err)
	}
	if err := os.Rename(tmpPath, finalDir); err != nil {
		os.Remove(tmpPath)
		return "", "", fmt.Errorf("rename to final: %w", err)
	}

	return actual, StoragePath(actual, s.filesDir), nil
}

// HashFile computes the SHA-256 of an existing file at the hash path.
func (s *Local) HashFile(fileHash string) (string, error) {
	if !ValidateHash(fileHash) {
		return "", fmt.Errorf("invalid file hash")
	}
	f, err := os.Open(AbsPath(s.root, fileHash, s.filesDir))
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Exists checks if the physical file exists on disk.
func (s *Local) Exists(fileHash string) bool {
	if !ValidateHash(fileHash) {
		return false
	}
	_, err := os.Stat(AbsPath(s.root, fileHash, s.filesDir))
	return err == nil
}

// Open opens the physical file for reading.
func (s *Local) Open(fileHash string) (*os.File, error) {
	if !ValidateHash(fileHash) {
		return nil, fmt.Errorf("invalid file hash")
	}
	return os.Open(AbsPath(s.root, fileHash, s.filesDir))
}

// ReadAt reads length bytes from the file at the given offset.
func (s *Local) ReadAt(fileHash string, offset, length int64) ([]byte, error) {
	if !ValidateHash(fileHash) {
		return nil, fmt.Errorf("invalid file hash")
	}
	f, err := os.Open(AbsPath(s.root, fileHash, s.filesDir))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	buf := make([]byte, length)
	n, err := f.ReadAt(buf, offset)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read at: %w", err)
	}
	return buf[:n], nil
}

// Delete removes the physical file from disk.
func (s *Local) Delete(fileHash string) error {
	if !ValidateHash(fileHash) {
		return fmt.Errorf("invalid file hash")
	}
	path := AbsPath(s.root, fileHash, s.filesDir)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

// CleanupUpload removes the temporary upload directory.
func (s *Local) CleanupUpload(uploadSlug string) error {
	if !ValidateSlug(uploadSlug) {
		return fmt.Errorf("invalid upload slug")
	}
	dir := filepath.Join(s.root, s.tmpDir, uploadSlug)
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("cleanup upload: %w", err)
	}
	return nil
}

// AbsPath returns the full absolute path for a file hash.
func (s *Local) AbsPath(fileHash string) string {
	return AbsPath(s.root, fileHash, s.filesDir)
}
