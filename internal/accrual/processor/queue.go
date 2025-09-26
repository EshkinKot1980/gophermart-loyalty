package processor

import (
	"context"
	"sync"
	"time"

	"github.com/EshkinKot1980/gophermart-loyalty/internal/accrual/config"
)

type Producer interface {
	ListToProccess(ctx context.Context) []string
}

type Consumer interface {
	Consume(ctx context.Context, sleepAll chan<- time.Duration, number string)
}

type MessageBroker struct {
	producer    Producer
	consumer    Consumer
	config      config.Config
	queueIn     chan string
	queueOut    <-chan string
	sleepSignal chan time.Duration
	haltSignal  chan struct{}
	stopped     chan struct{}
	mainCtx     context.Context
	mainCancel  context.CancelFunc
}

func NewMessageBroker(p Producer, c Consumer, cfg config.Config) *MessageBroker {
	b := &MessageBroker{
		producer:    p,
		consumer:    c,
		config:      cfg,
		queueIn:     make(chan string, int(cfg.RateLimit)),
		queueOut:    make(chan string),
		sleepSignal: make(chan time.Duration, 1),
	}

	b.mainCtx, b.mainCancel = context.WithCancel(context.Background())
	q := newQueue(b.queueIn, b.sleepSignal)
	b.queueOut = q.Out()

	return b
}

func (b *MessageBroker) Run(shutdownCxt context.Context) {
	go b.shutdownHandler(shutdownCxt)
	go b.runConsumers()
	time.Sleep(10 * time.Millisecond)
	go b.process()
}

func (b *MessageBroker) Stop() {
	select {
	case <-b.mainCtx.Done():
		time.Sleep(100 * time.Millisecond)
	case <-b.stopped:
	}

}

func (b *MessageBroker) shutdownHandler(ctx context.Context) {
	b.haltSignal = make(chan struct{})

	<-ctx.Done()
	close(b.haltSignal)
	close(b.queueIn)

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-b.stopped:
	case <-timeoutCtx.Done():
		b.mainCancel()
	}
}

func (b *MessageBroker) process() {
	produseInterval := time.Duration(b.config.PollInterval) * time.Second

	for {
		select {
		case <-b.haltSignal:
			return
		case <-time.After(produseInterval):
		}

		numbers := b.producer.ListToProccess(b.mainCtx)
		for _, number := range numbers {
			select {
			case <-b.haltSignal:
				return
			case b.queueIn <- number:
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func (b *MessageBroker) runConsumers() {
	var wg sync.WaitGroup
	count := int(b.config.RateLimit)
	wg.Add(count)
	b.stopped = make(chan struct{})

	for range count {
		go func() {
			b.consume()
			wg.Done()
		}()
	}

	wg.Wait()
	close(b.stopped)
}

func (b *MessageBroker) consume() {
	for {
		select {
		case <-b.haltSignal:
			return
		case number := <-b.queueOut:
			b.consumer.Consume(b.mainCtx, b.sleepSignal, number)
		case <-time.After(time.Millisecond):
		}
	}
}

type queue struct {
	in    <-chan string
	out   chan string
	sleep <-chan time.Duration
	mx    *sync.RWMutex
}

func newQueue(in <-chan string, sleep <-chan time.Duration) *queue {
	q := &queue{
		in:    in,
		out:   make(chan string),
		sleep: sleep,
		mx:    &sync.RWMutex{},
	}

	go q.sleepHandler()
	go q.process()
	time.Sleep(10 * time.Millisecond)

	return q
}

func (q *queue) Out() <-chan string {
	return q.out
}

func (q *queue) sleepHandler() {
	q.mx.Lock()
	until := time.Now()
	interval := 10 * time.Millisecond
	for {
		now := time.Now()
		select {
		case interval = <-q.sleep:
			u := now.Add(interval)
			if until.Before(u) {
				until = u
			}

			if q.mx.TryRLock() {
				q.mx.RUnlock()
				q.mx.Lock()
			}
		case <-time.After(interval):
			if now.Before(until) {
				interval = until.Sub(now)
			} else {
				interval = 10 * time.Millisecond

				if q.mx.TryRLock() {
					q.mx.RUnlock()
				} else {
					q.mx.Unlock()
				}
			}
		}
	}
}

func (q *queue) process() {
	for {
		select {
		case <-time.After(10 * time.Millisecond):
		case msg, ok := <-q.in:
			if !ok {
				return
			}

			q.mx.RLock()
			q.out <- msg
			q.mx.RUnlock()
		}
	}
}
