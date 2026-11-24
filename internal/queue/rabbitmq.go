package queue

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitQueue struct {
	conn *amqp.Connection
	pub  *amqp.Channel
	mu   sync.Mutex
}

type retryTier struct {
	Key string
	TTL time.Duration
}

const (
	exchange = "jobs"
	workQ    = "jobs.work"
	retryQ   = "jobs.retry"
	deadQ    = "jobs.dead"

	keyWork  = "work"
	keyRetry = "retry"
	keyDead  = "dead"
)

var retryTiers = []retryTier{
	{Key: "retry.5s", TTL: time.Second * 5},
	{Key: "retry.30s", TTL: time.Second * 30},
	{Key: "retry.2m", TTL: time.Minute * 2},
	{Key: "retry.10m", TTL: time.Minute * 10},
}

func NewRabbitQueue(url string) (*RabbitQueue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

}

func declareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(exchange, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	queues := []struct {
		name, key string
		args      amqp.Table
	}{
		{workQ, keyWork, amqp.Table{
			"x-dead-letter-exchange":    exchange,
			"x-dead-letter-routing-key": keyDead,
		}},
		{deadQ, keyDead, nil},
	}
	for _, q := range queues {
		if _, err := ch.QueueDeclare(q.name, true, false, false, false, q.args); err != nil {
			return err
		}

		if err := ch.QueueBind(q.name, q.key, exchange, false, nil); err != nil {
			return err
		}
	}

	for _, t := range retryTiers {
		args := amqp.Table{
			"x-message-ttl":             t.TTL.Milliseconds(),
			"x-dead-letter-exchange":    exchange,
			"x-dead-letter-routing-key": keyWork,
		}

		name := "jobs." + t.Key
		if _, err := ch.QueueDeclare(name, true, false, false, false, args); err != nil {
			return err
		}

		if err := ch.QueueBind(name, t.Key, exchange, false, nil); err != nil {
			return err
		}
	}
	return nil
}

func (q *RabbitQueue) Publish(ctx context.Context, jobID uuid.UUID) error {
	return q.publishToQueue(ctx, keyWork, jobID)
}

func (q *RabbitQueue) PublishDelayed(ctx context.Context, jobID uuid.UUID, attempt int) error {
	tier := retryTiers[min(attempt, len(retryTiers)-1)]
	return q.publishToQueue(ctx, tier.Key, jobID)
}

func (q *RabbitQueue) PublishDead(ctx context.Context, jobID uuid.UUID) error {
	return q.publishToQueue(ctx, keyDead, jobID)
}

func (q *RabbitQueue) publishToQueue(ctx context.Context, key string, jobID uuid.UUID) error {
	msg := amqp.Publishing{
		ContentType:  "text/plain",
		DeliveryMode: amqp.Persistent,
		MessageId:    jobID.string(),
		Body:         []byte(jobID.string()),
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	conf, err := q.pub.PublishWithDeferredConfirmWithContext(ctx, exchange, key, false, false, msg)
	if err != nil {
		return err
	}

	ok, err := conf.WaitContext(ctx)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("rabbitmq nacked publish")
	}

	return nil
}

func (q *RabbitQueue) Consume(ctx context.Context, concurrency int) error {
	ch, err := q.conn.Channel()
	if err != nil {
		return err
	}

	if err := ch.Qos(concurrency, 0, false); err != nil {
		return err
	}

	tag := "worker-" + uuid.NewString()
	deliveries, err := ch.Consume(workQ, tag, false, false, false, false, nil)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(){
			defer wg.Done()
			for d := deliveries {
				q.handle(ctx, d, h)
			}
		}()
	}

	<- ctx.Done()
	_ = ch.Cancel(tag, false)
	wg.Wait()
	return nil
}

func (q *RabbitQueue) Close() error {
	_ = q.pub.Close()
	return q.conn.Close()
}
