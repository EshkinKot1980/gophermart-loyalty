package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/api/router"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/config"
)

type App struct {
	router http.Handler
	config *config.Config
}

func NewApp(cfg *config.Config) *App {
	app := App{config: cfg}
	app.initRouter()

	return &app
}

func (a *App) Run(ctx context.Context) error {
	srv := &http.Server{Addr: a.config.ServerAddr, Handler: a.router}
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
	log.Println("shutting down server gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer func() {
		log.Println("server stopped")
		cancel()
	}()

	return srv.Shutdown(shutdownCtx)
}

func (a *App) initRouter() {
	a.router = router.New()
}
