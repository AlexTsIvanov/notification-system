package rabbitmq

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/AlexTsIvanov/notification-system/pkg/types"
	"github.com/streadway/amqp"
)

const (
	delayQueueFormat   = "%s.delay.%d"
	deadLetterExchange = "%s.dlx"
	deadLetterQueue    = "%s.dlq"

	initialTTL        = 1000
	exponentialFactor = 2
)

type RabbitMQBroker struct {
	conn                *amqp.Connection
	channel             *amqp.Channel
	mainQueueName       string
	deadLetterQueueName string
	maxRetries          int
	msgs                <-chan amqp.Delivery
}

func NewRabbitMQBroker(uri, mainQueueName string, maxRetry int, withConsumer bool) (*RabbitMQBroker, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	dlqName := fmt.Sprintf(deadLetterQueue, mainQueueName)
	dlxName := fmt.Sprintf(deadLetterExchange, mainQueueName)

	_, err = channel.QueueDeclare(dlqName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a DLQ: %v", err)
	}

	err = channel.ExchangeDeclare(dlxName, "direct", true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare DLX: %v", err)
	}

	err = channel.QueueBind(mainQueueName, mainQueueName, dlxName, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to bind main queue to DLX: %v", err)
	}

	_, err = channel.QueueDeclare(mainQueueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange": dlxName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to declare the main queue: %v", err)
	}

	for attempt := 1; attempt <= maxRetry; attempt++ {
		delayQueueName := fmt.Sprintf(delayQueueFormat, mainQueueName, attempt)
		ttl := initialTTL * int(math.Pow(float64(exponentialFactor), float64(attempt-1)))

		_, err = channel.QueueDeclare(delayQueueName, true, false, false, false, amqp.Table{
			"x-dead-letter-exchange":    dlxName,
			"x-message-ttl":             ttl,
			"x-dead-letter-routing-key": mainQueueName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to declare delay queue %s: %v", delayQueueName, err)
		}
	}

	var msgs <-chan amqp.Delivery
	if withConsumer {
		msgs, err = channel.Consume(
			mainQueueName,
			"",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to set up consumer: %v", err)
		}
	}

	return &RabbitMQBroker{
		conn:                conn,
		channel:             channel,
		mainQueueName:       mainQueueName,
		deadLetterQueueName: dlqName,
		maxRetries:          maxRetry,
		msgs:                msgs,
	}, nil
}

func (r *RabbitMQBroker) Close() {
	r.channel.Close()
	r.conn.Close()
}

func (r *RabbitMQBroker) Send(ctx context.Context, message []byte) error {
	err := r.channel.Publish(
		"",
		r.mainQueueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

func (r *RabbitMQBroker) Read(ctx context.Context) (types.EventContext, error) {
	if r.msgs == nil {
		return types.EventContext{}, fmt.Errorf("no consumer set up")
	}

	select {
	case d := <-r.msgs:
		headerValueStr := fmt.Sprintf("%v", d.Headers["x-retry-count"])

		retryCount, err := strconv.Atoi(headerValueStr)
		if err != nil {
			retryCount = 0
		}

		return types.EventContext{
			EventId:    fmt.Sprint(d.DeliveryTag),
			Payload:    d.Body,
			RetryCount: retryCount,
		}, nil
	case <-ctx.Done():
		return types.EventContext{}, ctx.Err()
	}
}

func (c *RabbitMQBroker) Ack(event types.EventContext) error {
	deliveryTag, err := strconv.ParseUint(event.EventId, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting id to int: %v", err)
	}

	return c.channel.Ack(deliveryTag, false)
}

func (c *RabbitMQBroker) Nack(event types.EventContext) error {
	deliveryTag, err := strconv.ParseUint(event.EventId, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting id to int: %v", err)
	}

	retryCount := event.RetryCount

	if retryCount >= c.maxRetries {

		err := c.channel.Publish(
			"",
			c.deadLetterQueueName,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				Body:         event.Payload,
				DeliveryMode: amqp.Persistent,
			},
		)
		if err != nil {
			// if we cannot publish the message to the delay queue
			// we Nack it and will be returned at the end of the queue it was
			return c.channel.Nack(deliveryTag, false, true)
		}

		return c.channel.Ack(deliveryTag, false)
	} else {
		delayQueueName := fmt.Sprintf(delayQueueFormat, c.mainQueueName, retryCount+1)

		retryHeaders := amqp.Table{
			"x-retry-count": retryCount + 1,
		}

		err := c.channel.Publish(
			"",
			delayQueueName,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				Body:         event.Payload,
				Headers:      retryHeaders,
				DeliveryMode: amqp.Persistent,
			},
		)
		if err != nil {
			// if we cannot publish the message to the delay queue
			// we Nack it and will be returned at the end of the queue it was
			return c.channel.Nack(deliveryTag, false, true)
		}

		return c.channel.Ack(deliveryTag, false)
	}
}
