package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexTsIvanov/notification-system/cmd/notification-service/env"
	"github.com/AlexTsIvanov/notification-system/cmd/notification-service/internal/consumer"
	"github.com/AlexTsIvanov/notification-system/cmd/notification-service/internal/factory"
	"github.com/AlexTsIvanov/notification-system/pkg/rabbitmq"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("loading application config...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := env.LoadAppConfig()
	if err != nil {
		logrus.Fatal("failed to load app config: ", err)
	}

	rabbitmqBroker, err := rabbitmq.NewRabbitMQBroker(config.RabbitMQUri, config.RabbitMQQueue, config.RabbitMQMaxRetries, true)
	if err != nil {
		logrus.Fatal("failed to init rabbitMQ broker: ", err)
	}
	defer rabbitmqBroker.Close()

	// TODO will probably need env vars for the different channels
	factory := factory.NewNotificationFactory()

	consumer := consumer.NewConsumer(rabbitmqBroker, factory)

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		<-signals
		logrus.Info("Received shutdown signal, exiting...")
		cancel()
	}()

	logrus.Info("entering consumer loop...")
	for {
		if err := consumer.HandleNotificationEvent(ctx); err != nil {
			logrus.Infof("Error consuming message: %v", err)
			if ctx.Err() == context.Canceled {
				break
			}
		}
	}
}
