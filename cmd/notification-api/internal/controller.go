package notification

import (
	"context"
	"encoding/json"
	"fmt"
)

type MessageBroker interface {
	Send(ctx context.Context, message []byte) error
}

type NotificationController struct {
	broker MessageBroker
}

func NewNotificationController(broker MessageBroker) *NotificationController {
	return &NotificationController{
		broker: broker,
	}
}

func (c *NotificationController) SendNotification(ctx context.Context, notification Notification) error {
	msg, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error marshaling event body: %v", err)
	}

	err = c.broker.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("error sending notification: %v", err)
	}
	return nil
}
