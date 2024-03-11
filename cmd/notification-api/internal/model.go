package notification

type NotificationRequest struct {
	Channel  string `json:"channel" validate:"required"`
	Content  string `json:"content" validate:"required"`
	Receiver string `json:"receiver" validate:"required"`
}