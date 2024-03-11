package notification

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type Controller interface {
	SendNotification(ctx context.Context, notification Notification) error
}

type Validator interface {
	Struct(s interface{}) error
}

type NotificationPresenter struct {
	contoller Controller
	validator Validator
}

func NewNotificationPresenter(cont Controller, validator Validator) *NotificationPresenter {
	return &NotificationPresenter{
		contoller: cont,
		validator: validator,
	}
}

func (p *NotificationPresenter) HandleSendNotification(c echo.Context) error {
	var request NotificationRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&request); err != nil {
		logrus.Errorf("failed to decode body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read request body")
	}

	// TODO improve validation from refined requirements
	if err := p.validator.Struct(request); err != nil {
		logrus.Errorf("failed to validate body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to validate body")
	}

	notification := Notification{
		Channel:  request.Channel,
		Content:  request.Content,
		Receiver: request.Receiver,
	}
	if err := p.contoller.SendNotification(c.Request().Context(), notification); err != nil {
		logrus.Errorf("failed to send notification: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to send notification")
	}

	return c.JSON(http.StatusAccepted, "notification request accepted")
}
