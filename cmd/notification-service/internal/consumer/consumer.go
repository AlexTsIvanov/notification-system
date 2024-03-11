package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AlexTsIvanov/notification-system/cmd/notification-service/internal/factory"
	"github.com/AlexTsIvanov/notification-system/pkg/types"
	"github.com/sirupsen/logrus"
)

type Reader interface {
	Read(ctx context.Context) (event types.EventContext, err error)
	Ack(event types.EventContext) error
	Nack(event types.EventContext) error
}

type Factory interface {
	GetSender(channel string) (factory.Sender, error)
}

type Consumer struct {
	reader  Reader
	factory Factory
}

func NewConsumer(reader Reader, factory Factory) *Consumer {
	return &Consumer{
		reader:  reader,
		factory: factory,
	}
}

func (c *Consumer) HandleNotificationEvent(ctx context.Context) error {
	event, err := c.reader.Read(ctx)
	if err != nil {
		return fmt.Errorf("error reading event queue: %v", err)
	}
	defer func() {
		if err != nil {
			if nackErr := c.reader.Nack(event); nackErr != nil {
				err = fmt.Errorf("%w; 2nd error: error sending negative acknowledgement: %v", err, nackErr)
			}
		} else {
			// here we have already sent the notification, not acking means event can be consumed twice
			// but that is acceptable from our requirements
			if ackErr := c.reader.Ack(event); ackErr != nil {
				logrus.Errorf("error sending acknowledgement: %v", ackErr)
			}
		}
	}()

	var notification Notification
	err = json.Unmarshal(event.Payload, &notification)
	if err != nil {
		return fmt.Errorf("error unmarshaling event body: %v", err)
	}

	channel, err := c.factory.GetSender(notification.Channel)
	if err != nil {
		return fmt.Errorf("error getting channel: %v", err)
	}

	// TODO - maybe the notification.Receiver is some userId and db has to be queried
	// to retrieve the channel specific receiver (email, phone for sms, slack id etc.)
	err = channel.Send(notification.Content, notification.Receiver)
	if err != nil {
		return fmt.Errorf("error sending notification: %v", err)
	}

	return nil
}
