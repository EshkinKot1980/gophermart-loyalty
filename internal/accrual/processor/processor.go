package processor

import (
	"context"
	"log"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/config"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/logger"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/repository/pg"
	"github.com/EshkinKot1980/gophermart-loyalty/internal/service"
)

type Processor struct {
	config *config.Config
	db     *pg.DB
	logger *logger.Logger
	queue  *MessageBroker
}

func New(c *config.Config, db *pg.DB, l *logger.Logger) *Processor {
	return &Processor{config: c, db: db, logger: l}
}

func (p *Processor) Run(ctx context.Context) {
	processRepository := repository.NewProcessing(
		p.db,
		p.config.ProcessDelay,
		p.config.UnregisteredRetries,
	)
	processService := service.NewProcessing(processRepository, p.logger)
	consumer := NewOrderConsumer(processService, p.logger, p.config.AccrualAddr)

	p.queue = NewMessageBroker(processService, consumer, *p.config)
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
