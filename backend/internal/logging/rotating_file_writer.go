package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type rotatingFileWriter struct {
	mu      sync.Mutex
	path    string
	maxSize int64
	file    *os.File
	size    int64
}

func newRotatingFileWriter(path string, maxSize int64) (*rotatingFileWriter, error) {
	if maxSize <= 0 {
		maxSize = int64(defaultMaxSizeMB) * 1024 * 1024
	}
	writer := &rotatingFileWriter{path: path, maxSize: maxSize}
	if err := writer.open(); err != nil {
		return nil, err
	}
	return writer, nil
}

func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}
	if w.size > 0 && w.size+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *rotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	w.size = 0
	return err
}

func (w *rotatingFileWriter) open() error {
	if err := os.MkdirAll(filepath.Dir(w.path), 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	if stat, err := os.Stat(w.path); err == nil && stat.Size() >= w.maxSize {
		if err := os.Rename(w.path, rotatedPath(w.path)); err != nil {
			return fmt.Errorf("rotate existing log file: %w", err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat log file: %w", err)
	}

	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("stat opened log file: %w", err)
	}
	w.file = file
	w.size = stat.Size()
	return nil
}

func (w *rotatingFileWriter) rotate() error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("close log file before rotate: %w", err)
		}
		w.file = nil
	}
	if err := os.Rename(w.path, rotatedPath(w.path)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("rotate log file: %w", err)
	}
	return w.open()
}

func rotatedPath(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	timestamp := time.Now().Format("20060102-150405")

	for index := 0; ; index++ {
		suffix := timestamp
		if index > 0 {
			suffix = fmt.Sprintf("%s-%d", timestamp, index)
		}
		candidate := filepath.Join(dir, fmt.Sprintf("%s-%s%s", name, suffix, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
