package consumer

type Notification struct {
	Channel  string `json:"channel"`
	Content  string `json:"content"`
	Receiver string `json:"receiver"`
}