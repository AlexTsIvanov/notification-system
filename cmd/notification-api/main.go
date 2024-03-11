package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlexTsIvanov/notification-system/cmd/notification-api/env"
	notification "github.com/AlexTsIvanov/notification-system/cmd/notification-api/internal"
	"github.com/AlexTsIvanov/notification-system/pkg/rabbitmq"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("loading application config...")
	config, err := env.LoadAppConfig()
	if err != nil {
		logrus.Fatal("failed to load app config: ", err)
	}

	e := echo.New()

	rabbitmqBroker, err := rabbitmq.NewRabbitMQBroker(config.RabbitMQUri, config.RabbitMQQueue, config.RabbitMQMaxRetries, false)
	if err != nil {
		logrus.Fatal("failed to init rabbitMQ broker: ", err)
	}
	defer rabbitmqBroker.Close()

	structValidator := validator.New()
	controller := notification.NewNotificationController(rabbitmqBroker)

	presenter := notification.NewNotificationPresenter(controller, structValidator)

	// TODO auth middleware, rate limiter
	e.POST("/send", presenter.HandleSendNotification)

	// Start server
	go func() {
		if err := e.Start(fmt.Sprintf("%s:%d", config.Host, config.Port)); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("failed to start server: ", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigChan
	signal.Stop(sigChan)
	logrus.Info("http server is stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logrus.Fatal("failed to shutdown server", err)
	}
}
