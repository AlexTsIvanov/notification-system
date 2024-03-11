package slack

type SlackSender struct{}

func NewSlackSender() *SlackSender {
	return &SlackSender{}
}

func (e *SlackSender) Send(message, recipient string) error {
	// implement slack specific logic here
	return nil
}