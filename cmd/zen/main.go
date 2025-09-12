package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonathandaddia/zen/internal/cli"
	"github.com/jonathandaddia/zen/internal/config"
	"github.com/jonathandaddia/zen/internal/logging"
)

func main() {
	// Setup graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// Use basic logging if config fails to load
		logger := logging.NewBasic()
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := logging.New(cfg.LogLevel, cfg.LogFormat)

	// Execute CLI
	if err := cli.Execute(ctx, cfg, logger); err != nil {
		logger.Error("execution failed", "error", err)
		os.Exit(1)
	}
}
