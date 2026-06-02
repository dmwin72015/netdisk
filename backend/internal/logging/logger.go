package logging

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

const (
	defaultOutput    = "console"
	defaultFilePath  = "./logs/server.log"
	defaultMaxSizeMB = 5
)

type Options struct {
	Level     string
	Output    string
	FilePath  string
	MaxSizeMB int
}

// New builds the application's structured logger. Unrecognised levels fall
// back to InfoLevel so we never crash the process over a typo in config.
func New(level string) zerolog.Logger {
	logger, _, _ := NewWithOptions(Options{Level: level})
	return logger
}

// NewWithOptions builds the application's structured logger and optionally
// writes logs to a rotating file. The returned closer is non-nil for file logs.
func NewWithOptions(opts Options) (zerolog.Logger, io.Closer, error) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	lvl, err := zerolog.ParseLevel(opts.Level)
	if err != nil || lvl == zerolog.NoLevel {
		lvl = zerolog.InfoLevel
	}

	writer, closer, err := buildWriter(opts)
	if err != nil {
		return zerolog.Logger{}, nil, err
	}

	return zerolog.New(writer).Level(lvl).With().Timestamp().Logger(), closer, nil
}

func buildWriter(opts Options) (io.Writer, io.Closer, error) {
	output := strings.ToLower(strings.TrimSpace(opts.Output))
	if output == "" {
		output = defaultOutput
	}

	switch output {
	case "console", "stdout", "terminal":
		return os.Stdout, nil, nil
	case "file":
		filePath := strings.TrimSpace(opts.FilePath)
		if filePath == "" {
			filePath = defaultFilePath
		}
		maxSizeMB := opts.MaxSizeMB
		if maxSizeMB <= 0 {
			maxSizeMB = defaultMaxSizeMB
		}
		writer, err := newRotatingFileWriter(filePath, int64(maxSizeMB)*1024*1024)
		if err != nil {
			return nil, nil, err
		}
		return writer, writer, nil
	default:
		return nil, nil, fmt.Errorf("unsupported log output %q", opts.Output)
	}
}
