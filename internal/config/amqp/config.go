package amqpconfig

import (
	"context"
	"fmt"
	"time"

	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpConfig struct {
	Host    string
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewAmqpConfig(host string) (*AmqpConfig, error) {
	conn, err := amqp.Dial(host)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to RabbitMQ: %s", err))
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Failed to open channel: %s", err)
	}

	err = ch.ExchangeDeclare(
		constants.EXCHANGE_NAME,
		amqp091.ExchangeDirect,
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare exchange: %s", err)
	}

	_, err = ch.QueueDeclare(
		constants.LIVE_LINES_QUEUE_NAME, // Name of the queue
		true,                            // Durable (persists across RabbitMQ restarts)
		false,                           // Delete when unused (automatically deleted when no consumers)
		false,                           // Exclusive (only one consumer can use the queue)
		true,                            // Wait (wait for the server to confirm the operation)
		nil,                             // Arguments (additional queue parameters)
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare queue: %s", err)
	}

	err = ch.QueueBind(
		constants.LIVE_LINES_QUEUE_NAME, // queue name
		constants.LIVE_LINES_QUEUE_NAME, // routing key
		constants.EXCHANGE_NAME,         // exchange
		true,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to bind a queue: %s", err)
	}

	return &AmqpConfig{
		Host:    host,
		Conn:    conn,
		Channel: ch,
	}, nil
}

func DefaultAmqpConfig() *AmqpConfig {
	config, err := NewAmqpConfig(utils.GetDotEnvVariable(constants.RABBITMQ_HOST))
	if err != nil {
		panic(fmt.Errorf("Failed to get amqp config: %s", err))
	}
	return config
}

func (c *AmqpConfig) DeclareAndBindQueue(queueName, routingKey string) error {
	_, err := c.Channel.QueueDeclare(
		queueName, // Name of the queue
		true,      // Durable (persists across RabbitMQ restarts)
		false,     // Delete when unused (automatically deleted when no consumers)
		false,     // Exclusive (only one consumer can use the queue)
		true,      // Wait (wait for the server to confirm the operation)
		nil,       // Arguments (additional queue parameters)
	)

	if err != nil {
		return err
	}

	err = c.Channel.QueueBind(
		routingKey,              // queue name
		routingKey,              // routing key
		constants.EXCHANGE_NAME, // exchange
		true,
		nil,
	)

	return err
}

func (c *AmqpConfig) PublishWithContext(data []byte, routingKey string, declareQueue bool) error {
	if declareQueue {
		err := c.DeclareAndBindQueue(routingKey, routingKey)
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Publish to the queue for the audit
	err := c.Channel.PublishWithContext(
		ctx,
		constants.EXCHANGE_NAME, // Exchange name
		routingKey,              // Routing key
		true,                    // Mandatory (if false, message is dropped if no queue is found)
		false,                   // Immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
	return err
}
