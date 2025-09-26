package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/processor"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/config"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/logger"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/pg"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := pg.NewDB(cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}
	defer db.Close()

	logger, err := logger.New()
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer logger.Sync()

	processor := processor.New(cfg.AccrualGfg, db, logger)
	processor.Run(ctx)
	defer processor.Stop()

	httpServer := api.NewApp(cfg, db, logger)
	return httpServer.Run(ctx)
}
