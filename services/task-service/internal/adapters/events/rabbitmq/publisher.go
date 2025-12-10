package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"

	"todoapp/pkg/events"
	"todoapp/services/task-service/internal/ports"
)

type Publisher struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
}

var _ ports.TaskEventPublisher = (*Publisher)(nil)

func New(url, queue string) (*Publisher, error) {
	if url == "" {
		return nil, errors.New("rabbitmq: url is required")
	}
	if queue == "" {
		return nil, errors.New("rabbitmq: queue is required")
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &Publisher{conn: conn, ch: ch, queue: queue}, nil
}

func (p *Publisher) Publish(ctx context.Context, event events.TaskEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx, "", p.queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         payload,
	})
}

func (p *Publisher) Close() error {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
