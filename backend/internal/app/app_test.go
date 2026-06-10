package app

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/netdisk/server/internal/config"
	"github.com/rs/zerolog"
)

func TestRunReturnsListenError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer listener.Close()

	a := &App{
		logger: zerolog.Nop(),
		srv:    &http.Server{Addr: listener.Addr().String()},
	}

	err = a.Run()
	if err == nil {
		t.Fatal("Run() error = nil, want listen error")
	}
	if !strings.Contains(err.Error(), "server listen") {
		t.Fatalf("Run() error = %v, want server listen error", err)
	}
}

func TestValidateStartupConfigRejectsInvalidCriticalValues(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{
			name: "nil config",
			cfg:  nil,
			want: "nil config",
		},
		{
			name: "invalid server port",
			cfg: &config.Config{
				Server: config.ServerConfig{Port: 0},
				Media:  config.MediaConfig{PollInterval: time.Second},
				Trash:  config.TrashConfig{PollInterval: time.Second},
			},
			want: "server.port",
		},
		{
			name: "invalid media interval",
			cfg: &config.Config{
				Server: config.ServerConfig{Port: 8080},
				Media:  config.MediaConfig{PollInterval: 0},
				Trash:  config.TrashConfig{PollInterval: time.Second},
			},
			want: "media.poll_interval",
		},
		{
			name: "invalid trash interval",
			cfg: &config.Config{
				Server: config.ServerConfig{Port: 8080},
				Media:  config.MediaConfig{PollInterval: time.Second},
				Trash:  config.TrashConfig{PollInterval: 0},
			},
			want: "trash.poll_interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStartupConfig(tt.cfg)
			if err == nil {
				t.Fatal("validateStartupConfig() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("validateStartupConfig() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestCriticalWorkerPanicReportsError(t *testing.T) {
	errCh := make(chan error, 1)
	app := &App{}

	app.startCriticalWorker(context.Background(), errCh, "test worker", func(context.Context) {
		panic("boom")
	})

	select {
	case err := <-errCh:
		if err == nil || !strings.Contains(err.Error(), "test worker panic: boom") {
			t.Fatalf("worker error = %v, want panic error", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for worker panic")
	}
}

func TestCriticalWorkerUnexpectedStopReportsError(t *testing.T) {
	errCh := make(chan error, 1)
	app := &App{}

	app.startCriticalWorker(context.Background(), errCh, "test worker", func(context.Context) {})

	select {
	case err := <-errCh:
		if err == nil || !strings.Contains(err.Error(), "test worker stopped unexpectedly") {
			t.Fatalf("worker error = %v, want stopped unexpectedly error", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for worker stop")
	}
}

func TestWriteStartupReadyFileCreatesFile(t *testing.T) {
	readyPath := filepath.Join(t.TempDir(), "nested", "server.ready")
	t.Setenv(startupReadyFileEnv, readyPath)

	if err := writeStartupReadyFile(); err != nil {
		t.Fatalf("writeStartupReadyFile() error = %v", err)
	}

	data, err := os.ReadFile(readyPath)
	if err != nil {
		t.Fatalf("read ready file: %v", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		t.Fatal("ready file is empty")
	}
}

func TestShouldServeFrontendFallback(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "", want: true},
		{path: "files/all", want: true},
		{path: "s/share-slug", want: true},
		{path: "_app/immutable/entry/start.js", want: false},
		{path: "app/immutable/entry/start.js", want: false},
		{path: "favicon.ico", want: false},
		{path: "robots.txt", want: false},
		{path: "assets/logo.png", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := shouldServeFrontendFallback(tt.path); got != tt.want {
				t.Fatalf("shouldServeFrontendFallback(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
