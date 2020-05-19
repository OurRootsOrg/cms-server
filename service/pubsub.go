package service

import (
	"context"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/awssnssqs"
	_ "gocloud.dev/pubsub/rabbitpubsub"
)

// If using RabbitMQ, you must set RABBIT_SERVER_URL; e.g., amqp://guest:guest@localhost:5672/
// The RabbitMQ url is rabbit://<exchange-name>
// The AWS url is awssqs://<queue-arn>?region=<region>
// Shutdown(ctx) the topic when you're done with it
func OpenTopic(ctx context.Context, urlStr string) (*pubsub.Topic, error) {
	return pubsub.OpenTopic(ctx, urlStr)
}

// If using RabbitMQ, you must set RABBIT_SERVER_URL; e.g., amqp://guest:guest@localhost:5672/
// The RabbitMQ url is rabbit://<queue-name>
// The AWS url is awssqs://<queue-arn>?region=<region>
// Shutdown(ctx) the subscription when you're done with it, and ack() messages when you've processed them
func OpenSubscription(ctx context.Context, urlStr string) (*pubsub.Subscription, error) {
	return pubsub.OpenSubscription(ctx, urlStr)
}
