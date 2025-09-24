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

type Queue struct {
	producer    Producer
	consumer    Consumer
	config      config.Config
	queue       chan string
	sleepSignal chan time.Duration
	haltSignal  chan struct{}
	stopped     chan struct{}
	mainCtx     context.Context
	mainCancel  context.CancelFunc
}

func NewQueue(p Producer, c Consumer, cfg config.Config) *Queue {
	q := &Queue{
		producer:    p,
		consumer:    c,
		config:      cfg,
		queue:       make(chan string),
		sleepSignal: make(chan time.Duration, 1),
	}

	q.mainCtx, q.mainCancel = context.WithCancel(context.Background())

	return q
}

func (q *Queue) Run(shutdownCxt context.Context) {
	go q.shutdownHandler(shutdownCxt)
	go q.runConsumers()
	time.Sleep(10 * time.Millisecond)
	go q.process()
}

func (q *Queue) Stop() {
	select {
	case <-q.mainCtx.Done():
		time.Sleep(100 * time.Millisecond)
	case <-q.stopped:
	}

}

func (q *Queue) shutdownHandler(ctx context.Context) {
	q.haltSignal = make(chan struct{})

	<-ctx.Done()
	close(q.haltSignal)
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-q.stopped:
	case <-timeoutCtx.Done():
		q.mainCancel()
	}
}

func (q *Queue) process() {
	produseInterval := time.Duration(q.config.PollInterval) * time.Second

	for {
		select {
		case <-q.haltSignal:
			return
		case <-time.After(produseInterval):
		}

		numbers := q.producer.ListToProccess(q.mainCtx)
		for _, number := range numbers {
			select {
			case <-q.haltSignal:
				return
			case d := <-q.sleepSignal:
				time.Sleep(d)
			default:
			}

			q.queue <- number
		}
		time.Sleep(time.Millisecond)
	}
}

func (q *Queue) runConsumers() {
	var wg sync.WaitGroup
	consumersCount := q.config.RateLimit
	wg.Add(int(consumersCount))
	q.stopped = make(chan struct{})

	for range consumersCount {
		go func() {
			q.consume()
			wg.Done()
		}()
	}

	wg.Wait()
	close(q.stopped)
}

func (q *Queue) consume() {
	var number string
	for {
		select {
		case <-q.haltSignal:
			return
		case number = <-q.queue:
		}

		q.consumer.Consume(q.mainCtx, q.sleepSignal, number)
		time.Sleep(10 * time.Millisecond)
	}
}
