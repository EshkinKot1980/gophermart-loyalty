package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/processor"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/api"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/config"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/logger"
)

func main() {
	cfg := config.MustLoad()

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg *config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	m, err := migrate.New("file://db/migrations", cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations to the DB: %w", err)
	}

	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("failed to create DB a connection pool: %w", err)
	}
	defer dbPool.Close()

	logger, err := logger.New()
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer logger.Sync()

	processor := processor.New(cfg.AccrualGfg, dbPool, logger)
	processor.Run(ctx)
	defer processor.Stop()

	httpServer := api.NewApp(cfg, dbPool, logger)
	return httpServer.Run(ctx)
}
