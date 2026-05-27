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

	logger := logging.New(cfg.Log.Level)

	a, err := app.New(context.Background(), cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("app init")
	}
	if err := a.Run(); err != nil {
		logger.Fatal().Err(err).Msg("app run")
	}
}
