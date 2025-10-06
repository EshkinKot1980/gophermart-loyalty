package processor

import (
	"context"
	"sync"
	"time"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/config"
)

type Producer interface {
	ListForProccess(ctx context.Context) []string
}

type Consumer interface {
	Consume(ctx context.Context, number string) (retryAfter time.Duration)
}

type MessageBroker struct {
	producer   Producer
	consumer   Consumer
	config     config.Config
	sleepUntil time.Time
	sleepMx    *sync.RWMutex
	queue      chan string
	halt       chan struct{}
	stopped    chan struct{}
}

func NewMessageBroker(p Producer, c Consumer, cfg config.Config) *MessageBroker {
	return &MessageBroker{
		producer:   p,
		consumer:   c,
		config:     cfg,
		sleepUntil: time.Now(),
		sleepMx:    &sync.RWMutex{},
		queue:      make(chan string, int(cfg.RateLimit)),
		halt:       make(chan struct{}),
		stopped:    make(chan struct{}),
	}
}

func (b *MessageBroker) Run(shutdownCxt context.Context) {
	mainCtx := b.runShutdownHandler(shutdownCxt)
	go b.runConsumers(mainCtx)
	go b.process(mainCtx)
}

func (b *MessageBroker) Stop() {
	<-b.stopped
}

func (b *MessageBroker) runShutdownHandler(ctx context.Context) context.Context {
	mainCtx, cancel := context.WithCancel(context.Background())

	go func() {
		<-ctx.Done()
		close(b.halt)
		timeoutCtx, c := context.WithTimeout(context.Background(), 3*time.Second)
		defer c()

		<-timeoutCtx.Done()
		cancel()
	}()

	return mainCtx
}

func (b *MessageBroker) process(ctx context.Context) {
	defer close(b.queue)
	produseInterval := time.Duration(b.config.PollInterval) * time.Second

	for {
		select {
		case <-b.halt:
			return
		case <-time.After(produseInterval):
		}

		numbers := b.producer.ListForProccess(ctx)
		for _, number := range numbers {
			select {
			case <-b.halt:
				return
			default:
			}
			b.queue <- number
		}
	}
}

func (b *MessageBroker) runConsumers(ctx context.Context) {
	defer close(b.stopped)
	var (
		wg    sync.WaitGroup
		count = int(b.config.RateLimit)
	)
	wg.Add(count)
	for range count {
		go func() {
			defer wg.Done()
			b.consume(ctx)
		}()
	}
	wg.Wait()
}

func (b *MessageBroker) consume(ctx context.Context) {
	for number := range b.queue {
		if !b.wakeupConsumer() {
			return
		}
		retryAfter := b.consumer.Consume(ctx, number)
		if retryAfter > 0 {
			b.sleepConsumers(retryAfter)
		}
	}
}

func (b *MessageBroker) wakeupConsumer() bool {
	var interval time.Duration
	for {
		select {
		case <-b.halt:
			return false
		case <-time.After(interval):
			b.sleepMx.RLock()
			until := b.sleepUntil
			b.sleepMx.RUnlock()

			now := time.Now()
			if now.After(until) {
				return true
			}
			interval = until.Sub(now)
		}
	}
}

func (b *MessageBroker) sleepConsumers(d time.Duration) {
	b.sleepMx.Lock()
	defer b.sleepMx.Unlock()

	until := time.Now().Add(d)
	if b.sleepUntil.Before(until) {
		b.sleepUntil = until
	}
}
