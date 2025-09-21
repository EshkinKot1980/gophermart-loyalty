package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/router"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/config"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/logger"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service"
)

type App struct {
	config *config.Config
	logger *logger.Logger
	dbPool *pgxpool.Pool
}

func NewApp(c *config.Config, p *pgxpool.Pool, l *logger.Logger) *App {
	return &App{config: c, dbPool: p, logger: l}
}

func (a *App) Run(ctx context.Context) error {
	srv := &http.Server{Addr: a.config.ServerAddr, Handler: a.newRouter()}
	errChan := make(chan error)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(time.Second):
		log.Printf("server listening on %s\n", a.config.ServerAddr)
	}

	<-ctx.Done()
	log.Println("shutting down http server gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer func() {
		log.Println("http server stopped")
		cancel()
	}()

	return srv.Shutdown(shutdownCtx)
}

func (a *App) newRouter() http.Handler {
	userRepository := repository.NewUser(a.dbPool)
	orderRepository := repository.NewOrder(a.dbPool)
	balanceRepository := repository.NewBalance(a.dbPool)
	withdrawalsRepository := repository.NewWithdrawals(a.dbPool)

	authService := service.NewAuth(userRepository, a.logger, a.config.JWTsecret)
	orderService := service.NewOrder(orderRepository, a.logger)
	balanceService := service.NewBalance(balanceRepository, a.logger)
	withdrawalsService := service.NewWithdrawals(withdrawalsRepository, a.logger)

	return router.New(
		authService,
		orderService,
		balanceService,
		withdrawalsService,
		a.logger,
	)
}
