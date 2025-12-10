package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"todoapp/services/notification-service/internal/domain/events"
	"todoapp/services/notification-service/internal/ports"
)

type Consumer struct {
	url      string
	queue    string
	logger   *log.Logger
	prefetch int
}

func NewConsumer(url, queue string, logger *log.Logger) *Consumer {
	return &Consumer{
		url:      url,
		queue:    queue,
		logger:   logger,
		prefetch: 10,
	}
}

func (c *Consumer) Consume(ctx context.Context, handler ports.EventHandler) error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if _, err := ch.QueueDeclare(c.queue, true, false, false, false, nil); err != nil {
		return err
	}

	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(c.queue, "notification-service", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("rabbitmq: deliveries channel closed")
			}
			if err := c.handleMessage(ctx, handler, &msg); err != nil {
				c.logger.Printf("failed to process message: %v", err)
				_ = msg.Nack(false, false)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, handler ports.EventHandler, msg *amqp.Delivery) error {
	var event events.TaskEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return handler.Handle(ctx, event)
}
