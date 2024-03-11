# Notification System

This system is designed to enable notifications to be sent across various channels including email, SMS, and Slack, with an architecture that supports easy extensibility for additional channels. Built for horizontal scalability, this system guarantees an "at least once" delivery SLA for messages.

## System Architecture

The Notification System comprises two main services:

- **Notification-API**: An HTTP service that accepts notification requests and publishes them to a RabbitMQ exchange.
- **Notification-Service**: A consumer service that listens to RabbitMQ events and uses a factory pattern to handle notification dispatch across different channels.

## Getting Started

### Prerequisites

Before setting up the Notification Gateway, ensure you have the following installed:
- Go
- RabbitMQ Server

### Setup Instructions

#### Setting Up RabbitMQ

For Windows (using PowerShell):
To start RabbitMQ on Windows, you can use the following command in PowerShell:

```powershell
rabbitmq-service start
```

For macOS:
To start RabbitMQ on macOS, open the Terminal and run the following command:

```bash
rabbitmq-server start
```

#### Starting the Go services

To start the notification-api:

```
go run .\cmd\notification-api\main.go
```

To start the notification-service:

```
go run .\cmd\notification-service\main.go
```

#### Sending Notifications
To send a notification, make an HTTP POST request to the notification-api send url(default is http://localhost:8080/send) with the following payload:

```json
{
  "channel": "email",
  "content": "Hello, this is a test notification!",
  "recipient": "user@example.com"
}
```

#### Extending the System

To add support for new notification channels:

1.Implement a new sender in the notification-service that conforms to the existing sender interface.

2.Update the factory function to recognize and use the new sender based on the notification request.