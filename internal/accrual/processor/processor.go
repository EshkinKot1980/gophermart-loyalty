package processor

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/config"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/logger"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service"
)

type Processor struct {
	config *config.Config
	dbPool *pgxpool.Pool
	logger *logger.Logger
	queue  *Queue
}

func New(c *config.Config, p *pgxpool.Pool, l *logger.Logger) *Processor {
	return &Processor{config: c, dbPool: p, logger: l}
}

func (p *Processor) Run(ctx context.Context) {
	processRepository := repository.NewProcessing(
		p.dbPool,
		p.config.ProcessDelay,
		p.config.UnregisteredRetries,
	)
	processService := service.NewProcessing(processRepository, p.logger)
	consumer := NewOrderConsumer(processService, p.logger, p.config.AccrualAddr)

	p.queue = NewQueue(processService, consumer, *p.config)
	p.queue.Run(ctx)

	go func() {
		<-ctx.Done()
		log.Println("shutting down acciral processor gracefully")
	}()
	log.Println("accruel processor started")
}

func (p *Processor) Stop() {
	p.queue.Stop()
	log.Println("acciral processor stopped")
}
