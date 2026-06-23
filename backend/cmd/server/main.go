package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/netdisk/server/internal/app"
	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/logging"
	"github.com/rs/zerolog"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config.yaml")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	logger, closer, err := logging.NewWithOptions(logging.Options{
		Level:     cfg.Log.Level,
		Output:    cfg.Log.Output,
		FilePath:  cfg.Log.FilePath,
		MaxSizeMB: cfg.Log.MaxSizeMB,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	if closer != nil {
		defer closer.Close()
	}

	a, err := app.New(context.Background(), cfg, logger)
	if err != nil {
		exitWithLoggedError(logger, closer, "app init", err)
	}
	if err := a.Run(); err != nil {
		exitWithLoggedError(logger, closer, "app run", err)
	}
}

func exitWithLoggedError(logger zerolog.Logger, closer io.Closer, msg string, err error) {
	logger.Error().Err(err).Msg(msg)
	fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	if closer != nil {
		_ = closer.Close()
	}
	os.Exit(1)
}
