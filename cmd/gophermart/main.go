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

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/config"
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
	m.Up() // TODO: как только в появятся миграции включить обработку ошибок
	// if err := m.Up(); err != nil {
	// 	return fmt.Errorf("failed to apply migrations to the DB: %w", err)
	// }

	httpServer := api.NewApp(cfg)
	return httpServer.Run(ctx)
}
