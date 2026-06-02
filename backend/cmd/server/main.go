package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/netdisk/server/internal/app"
	"github.com/netdisk/server/internal/config"
	"github.com/netdisk/server/internal/logging"
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
		logger.Fatal().Err(err).Msg("app init")
	}
	if err := a.Run(); err != nil {
		logger.Fatal().Err(err).Msg("app run")
	}
}
