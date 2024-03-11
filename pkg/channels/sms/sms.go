package sms

type SMSSender struct{}

func NewSMSSender() *SMSSender {
	return &SMSSender{}
}

func (e *SMSSender) Send(message, recipient string) error {
	// implement sms specific logic here
	// maybe an sms template is needed that needs to be fetched from db
	// if used a lot maybe templates can be stored in a cache
	return nil
}