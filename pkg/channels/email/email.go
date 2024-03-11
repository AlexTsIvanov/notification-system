package email

import "github.com/sirupsen/logrus"

type EmailSender struct{}

func NewEmailSender() *EmailSender {
	return &EmailSender{}
}

func (e *EmailSender) Send(message, recipient string) error {
	// implement email specific logic here
	// maybe an email template is needed that needs to be fetched from db
	// if used a lot maybe templates can be stored in a cache

	logrus.Info(message, recipient)
	return nil
}
