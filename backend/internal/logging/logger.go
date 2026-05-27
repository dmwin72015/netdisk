package logging

import (
	"os"

	"github.com/rs/zerolog"
)

// New builds the application's structured logger. Unrecognised levels fall
// back to InfoLevel so we never crash the process over a typo in config.
func New(level string) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	lvl, err := zerolog.ParseLevel(level)
	if err != nil || lvl == zerolog.NoLevel {
		lvl = zerolog.InfoLevel
	}
	return zerolog.New(os.Stdout).Level(lvl).With().Timestamp().Logger()
}
