package factory

import (
	"fmt"

	"github.com/AlexTsIvanov/notification-system/pkg/channels/email"
	"github.com/AlexTsIvanov/notification-system/pkg/channels/slack"
	"github.com/AlexTsIvanov/notification-system/pkg/channels/sms"
)


//go:generate mockgen --source=factory.go --destination ../mocks/factory.go --package mocks

// needs to be implemented by all sending channels
type Sender interface {
	Send(message, recipient string) error
}

type NotificationFactory struct{}

func NewNotificationFactory() *NotificationFactory {
	return &NotificationFactory{}
}

func (f NotificationFactory) GetSender(channel string) (Sender, error) {
	switch channel {
	case "email":
		return email.NewEmailSender(), nil
	case "sms":
		return sms.NewSMSSender(), nil
	case "slack":
		return slack.NewSlackSender(), nil
	default:
		return nil, fmt.Errorf("Unsupported notification channel: %s", channel)
	}
}
