package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
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

// ChunkExists checks whether a chunk file exists on disk with the expected size.
func (s *Local) ChunkExists(uploadSlug string, chunkIndex int, expectedSize int64) bool {
	if !ValidateSlug(uploadSlug) {
		slog.Warn("chunk-exists: invalid slug", "slug", uploadSlug)
		return false
	}
	path := filepath.Join(s.root, s.tmpDir, uploadSlug, fmt.Sprintf("chunk_%06d", chunkIndex))
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("chunk-exists: stat failed", "slug", uploadSlug, "chunkIndex", chunkIndex, "err", err)
		}
		return false
	}
	exists := info.Size() == expectedSize
	slog.Debug("chunk-exists", "slug", uploadSlug, "chunkIndex", chunkIndex, "expectedSize", expectedSize, "actualSize", info.Size(), "exists", exists)
	return exists
}

// WriteChunk writes a single chunk to the temporary upload directory.
func (s *Local) WriteChunk(uploadSlug string, chunkIndex int, data io.Reader) error {
	if !ValidateSlug(uploadSlug) {
		return fmt.Errorf("invalid upload slug")
	}
	dir := filepath.Join(s.root, s.tmpDir, uploadSlug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Error("write-chunk: mkdir failed", "slug", uploadSlug, "chunkIndex", chunkIndex, "dir", dir, "err", err)
		return fmt.Errorf("mkdir chunk dir: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("chunk_%06d", chunkIndex))
	slog.Debug("write-chunk: creating", "slug", uploadSlug, "chunkIndex", chunkIndex, "path", path)
	f, err := os.Create(path)
	if err != nil {
		slog.Error("write-chunk: create file failed", "slug", uploadSlug, "chunkIndex", chunkIndex, "path", path, "err", err)
		return fmt.Errorf("create chunk: %w", err)
	}
	defer f.Close()
	written, err := io.Copy(f, data)
	if err != nil {
		slog.Error("write-chunk: io copy failed", "slug", uploadSlug, "chunkIndex", chunkIndex, "path", path, "err", err)
		return fmt.Errorf("write chunk: %w", err)
	}
	slog.Debug("write-chunk: success", "slug", uploadSlug, "chunkIndex", chunkIndex, "bytesWritten", written)
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
		slog.Warn("inspect-chunks: invalid metadata", "slug", uploadSlug, "totalChunks", totalChunks, "chunkSize", chunkSize, "fileSize", fileSize)
		return nil, nil, fmt.Errorf("invalid chunk metadata")
	}

	chunkDir := filepath.Join(s.root, s.tmpDir, uploadSlug)
	slog.Debug("inspect-chunks: start", "slug", uploadSlug, "totalChunks", totalChunks, "chunkDir", chunkDir)

	valid := make(map[int]bool)
	issues := make([]ChunkIssue, 0)
	missingCount := 0
	sizeMismatchCount := 0
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
				missingCount++
				continue
			}
			slog.Error("inspect-chunks: stat failed", "slug", uploadSlug, "chunkIndex", i, "path", chunkPath, "err", err)
			return nil, nil, fmt.Errorf("stat chunk %d: %w", i, err)
		}
		if info.Size() == expected {
			valid[i] = true
		} else {
			issues = append(issues, ChunkIssue{Index: i, Expected: expected, Actual: info.Size()})
			sizeMismatchCount++
		}
	}

	slog.Debug("inspect-chunks: complete", "slug", uploadSlug, "totalChunks", totalChunks, "validCount", len(valid), "missingCount", missingCount, "sizeMismatchCount", sizeMismatchCount)
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

	mergeStart := time.Now()
	slog.Info("merge-chunks: start", "slug", uploadSlug, "totalChunks", totalChunks, "hashPrefix", safePrefix(fileHash))

	finalDir := filepath.Join(s.root, StoragePath(fileHash, s.filesDir))
	if err := os.MkdirAll(filepath.Dir(finalDir), 0o755); err != nil {
		slog.Error("merge-chunks: mkdir final dir failed", "slug", uploadSlug, "finalDir", filepath.Dir(finalDir), "err", err)
		return "", fmt.Errorf("mkdir final dir: %w", err)
	}

	// Write to a temporary .merging file first.
	mergingPath := finalDir + ".merging"
	slog.Debug("merge-chunks: creating merging file", "slug", uploadSlug, "mergingPath", mergingPath)
	out, err := os.Create(mergingPath)
	if err != nil {
		slog.Error("merge-chunks: create merging file failed", "slug", uploadSlug, "mergingPath", mergingPath, "err", err)
		return "", fmt.Errorf("create merging file: %w", err)
	}

	h := sha256.New()
	writer := io.MultiWriter(out, h)

	chunkDir := filepath.Join(s.root, s.tmpDir, uploadSlug)
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%06d", i))
		chunkStart := time.Now()
		f, err := os.Open(chunkPath)
		if err != nil {
			out.Close()
			os.Remove(mergingPath)
			slog.Error("merge-chunks: open chunk failed", "slug", uploadSlug, "chunkIndex", i, "path", chunkPath, "err", err)
			return "", fmt.Errorf("open chunk %d: %w", i, err)
		}
		n, err := io.Copy(writer, f)
		if err != nil {
			f.Close()
			out.Close()
			os.Remove(mergingPath)
			slog.Error("merge-chunks: copy chunk failed", "slug", uploadSlug, "chunkIndex", i, "bytesCopied", n, "err", err)
			return "", fmt.Errorf("copy chunk %d: %w", i, err)
		}
		f.Close()
		slog.Debug("merge-chunks: chunk merged", "slug", uploadSlug, "chunkIndex", i, "bytesCopied", n, "duration", time.Since(chunkStart))
	}
	out.Close()

	// Verify SHA-256.
	actual := hex.EncodeToString(h.Sum(nil))
	if actual != fileHash {
		os.Remove(mergingPath)
		slog.Error("merge-chunks: hash mismatch", "slug", uploadSlug, "expected", fileHash, "actual", actual)
		return "", fmt.Errorf("hash mismatch: expected %s, got %s", fileHash, actual)
	}
	slog.Debug("merge-chunks: hash verified", "slug", uploadSlug)

	// Atomic rename to final path.
	if err := os.Rename(mergingPath, finalDir); err != nil {
		os.Remove(mergingPath)
		slog.Error("merge-chunks: rename failed", "slug", uploadSlug, "from", mergingPath, "to", finalDir, "err", err)
		return "", fmt.Errorf("rename to final: %w", err)
	}

	slog.Info("merge-chunks: complete", "slug", uploadSlug, "totalChunks", totalChunks, "duration", time.Since(mergeStart))
	return StoragePath(fileHash, s.filesDir), nil
}

// MergeChunksAndHash merges chunks, computes SHA-256, and returns the hash.
// Used when file hash is not known at init time (pre_hash fast path).
func (s *Local) MergeChunksAndHash(uploadSlug string, totalChunks int) (string, string, error) {
	if !ValidateSlug(uploadSlug) {
		return "", "", fmt.Errorf("invalid upload slug")
	}

	mergeStart := time.Now()
	slog.Info("merge-chunks-and-hash: start", "slug", uploadSlug, "totalChunks", totalChunks)

	chunkDir := filepath.Join(s.root, s.tmpDir, uploadSlug)

	// Merge to temp file and compute hash.
	tmpPath := filepath.Join(s.root, s.tmpDir, uploadSlug+".merged")
	slog.Debug("merge-chunks-and-hash: creating temp file", "slug", uploadSlug, "tmpPath", tmpPath)
	out, err := os.Create(tmpPath)
	if err != nil {
		slog.Error("merge-chunks-and-hash: create temp file failed", "slug", uploadSlug, "tmpPath", tmpPath, "err", err)
		return "", "", fmt.Errorf("create temp file: %w", err)
	}

	h := sha256.New()
	writer := io.MultiWriter(out, h)

	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%06d", i))
		chunkStart := time.Now()
		f, err := os.Open(chunkPath)
		if err != nil {
			out.Close()
			os.Remove(tmpPath)
			slog.Error("merge-chunks-and-hash: open chunk failed", "slug", uploadSlug, "chunkIndex", i, "path", chunkPath, "err", err)
			return "", "", fmt.Errorf("open chunk %d: %w", i, err)
		}
		n, err := io.Copy(writer, f)
		if err != nil {
			f.Close()
			out.Close()
			os.Remove(tmpPath)
			slog.Error("merge-chunks-and-hash: copy chunk failed", "slug", uploadSlug, "chunkIndex", i, "bytesCopied", n, "err", err)
			return "", "", fmt.Errorf("copy chunk %d: %w", i, err)
		}
		f.Close()
		slog.Debug("merge-chunks-and-hash: chunk merged", "slug", uploadSlug, "chunkIndex", i, "bytesCopied", n, "duration", time.Since(chunkStart))
	}
	out.Close()

	actual := hex.EncodeToString(h.Sum(nil))
	slog.Debug("merge-chunks-and-hash: computed hash", "slug", uploadSlug, "hashPrefix", safePrefix(actual))

	// Move to final path based on computed hash.
	finalDir := filepath.Join(s.root, StoragePath(actual, s.filesDir))
	if err := os.MkdirAll(filepath.Dir(finalDir), 0o755); err != nil {
		os.Remove(tmpPath)
		slog.Error("merge-chunks-and-hash: mkdir final dir failed", "slug", uploadSlug, "finalDir", filepath.Dir(finalDir), "err", err)
		return "", "", fmt.Errorf("mkdir final dir: %w", err)
	}
	if err := os.Rename(tmpPath, finalDir); err != nil {
		os.Remove(tmpPath)
		slog.Error("merge-chunks-and-hash: rename failed", "slug", uploadSlug, "from", tmpPath, "to", finalDir, "err", err)
		return "", "", fmt.Errorf("rename to final: %w", err)
	}

	slog.Info("merge-chunks-and-hash: complete", "slug", uploadSlug, "totalChunks", totalChunks, "hashPrefix", safePrefix(actual), "duration", time.Since(mergeStart))
	return actual, StoragePath(actual, s.filesDir), nil
}

func safePrefix(s string) string {
	if len(s) >= 16 {
		return s[:16] + "..."
	}
	return s
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
	slog.Debug("cleanup-upload: removing temp dir", "slug", uploadSlug, "dir", dir)
	if err := os.RemoveAll(dir); err != nil {
		slog.Warn("cleanup-upload: removeAll failed", "slug", uploadSlug, "dir", dir, "err", err)
		return fmt.Errorf("cleanup upload: %w", err)
	}
	slog.Debug("cleanup-upload: done", "slug", uploadSlug)
	return nil
}

// WriteFromReader streams data from r to a sharded file, computes SHA-256,
// and returns the hash, storage path, total bytes written, and preHash.
// Cleans up temp files on error.
func (s *Local) WriteFromReader(r io.Reader) (fileHash string, storagePath string, fileSize int64, preHash string, err error) {
	slug, nerr := gonanoid.New(21)
	if nerr != nil {
		return "", "", 0, "", fmt.Errorf("generate slug: %w", nerr)
	}

	tmpDir := filepath.Join(s.root, s.tmpDir, "dl_"+slug)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", "", 0, "", fmt.Errorf("mkdir temp: %w", err)
	}

	defer func() {
		if err != nil {
			os.RemoveAll(tmpDir)
		}
	}()

	tmpPath := filepath.Join(tmpDir, "download")
	f, cerr := os.Create(tmpPath)
	if cerr != nil {
		return "", "", 0, "", fmt.Errorf("create temp file: %w", cerr)
	}

	h := sha256.New()

	preHashBuf := make([]byte, 64*1024)
	n, readErr := io.ReadFull(r, preHashBuf)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) && !errors.Is(readErr, io.EOF) {
		f.Close()
		os.Remove(tmpPath)
		return "", "", 0, "", fmt.Errorf("read prehash: %w", readErr)
	}
	preHashData := preHashBuf[:n]
	ph := sha256.Sum256(preHashData)
	preHashHex := hex.EncodeToString(ph[:])

	writer := io.MultiWriter(f, h)
	if _, werr := writer.Write(preHashData); werr != nil {
		f.Close()
		os.Remove(tmpPath)
		return "", "", 0, "", fmt.Errorf("write first part: %w", werr)
	}

	copied, copyErr := io.Copy(writer, r)
	if copyErr != nil {
		f.Close()
		os.Remove(tmpPath)
		return "", "", 0, "", fmt.Errorf("copy stream: %w", copyErr)
	}
	f.Close()

	fileSize = int64(len(preHashData)) + copied
	fileHashVal := hex.EncodeToString(h.Sum(nil))

	// Move to final sharded path
	finalDir := filepath.Join(s.root, StoragePath(fileHashVal, s.filesDir))
	if err := os.MkdirAll(filepath.Dir(finalDir), 0o755); err != nil {
		os.Remove(tmpPath)
		return "", "", 0, "", fmt.Errorf("mkdir final: %w", err)
	}
	if err := os.Rename(tmpPath, finalDir); err != nil {
		os.Remove(tmpPath)
		return "", "", 0, "", fmt.Errorf("rename to final: %w", err)
	}

	os.RemoveAll(tmpDir)

	return fileHashVal, StoragePath(fileHashVal, s.filesDir), fileSize, preHashHex, nil
}

// AbsPath returns the full absolute path for a file hash.
func (s *Local) AbsPath(fileHash string) string {
	return AbsPath(s.root, fileHash, s.filesDir)
}
