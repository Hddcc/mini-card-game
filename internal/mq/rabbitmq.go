package mq

import (
	"context"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	exchange   string
	queue      string
	routingKey string
}

type Config struct {
	URL        string
	Exchange   string
	Queue      string
	RoutingKey string
}

func NewRabbitMQ(cfg Config) (*RabbitMQ, error) {
	if cfg.URL == "" {
		return nil, errors.New("rabbitmq url empty")
	}
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	client := &RabbitMQ{
		conn:       conn,
		channel:    ch,
		exchange:   cfg.Exchange,
		queue:      cfg.Queue,
		routingKey: cfg.RoutingKey,
	}
	if err := client.DeclareAwardTopology(); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

func (r *RabbitMQ) DeclareAwardTopology() error {
	if r == nil || r.channel == nil {
		return errors.New("rabbitmq not connected")
	}
	if err := r.channel.ExchangeDeclare(
		r.exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}
	if _, err := r.channel.QueueDeclare(
		r.queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}
	return r.channel.QueueBind(r.queue, r.routingKey, r.exchange, false, nil)
}

func (r *RabbitMQ) Publish(ctx context.Context, body []byte) error {
	if r == nil || r.channel == nil {
		return errors.New("rabbitmq not connected")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.channel.PublishWithContext(
		ctx,
		r.exchange,
		r.routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}

func (r *RabbitMQ) Consume(consumerTag string) (<-chan amqp.Delivery, error) {
	if r == nil || r.channel == nil {
		return nil, errors.New("rabbitmq not connected")
	}
	return r.channel.Consume(
		r.queue,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) Close() {
	if r == nil {
		return
	}
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}
