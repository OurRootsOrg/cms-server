package api

import (
	"context"
	"errors"
	"log"
	"math"
	"strings"
	"time"

	"github.com/streadway/amqp"
	"gocloud.dev/pubsub/rabbitpubsub"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/awssnssqs"
	_ "gocloud.dev/pubsub/rabbitpubsub"
)

// OpenTopic opens a topic for publishing
// Shutdown(ctx) the topic when you're done with it
func (api *API) OpenTopic(ctx context.Context, topicName string) (*pubsub.Topic, error) {
	cnt := 0
	err := errors.New("unknown error")
	urlStr := api.pubSubConfig.QueueURL(topicName)
	conn := api.rabbitmqTopicConn
	var topic *pubsub.Topic
	for err != nil && cnt <= 5 {
		if cnt > 0 {
			time.Sleep(time.Duration(math.Pow(2.0, float64(cnt))) * time.Second)
		}
		cnt++

		switch {
		case strings.HasPrefix(urlStr, "https://sqs."): // AWS SQS https URL
			urlStr = "awssqs" + urlStr[5:]
			topic, err = pubsub.OpenTopic(ctx, urlStr)
		case strings.HasPrefix(urlStr, "amqp:"): // Rabbit
			if conn == nil {
				conn, err = amqp.Dial(urlStr)
				if err != nil {
					log.Printf("[INFO] Rabbit dialed try %d error %v\n", cnt, err)
					conn = nil
					continue
				}
			}
			topic = rabbitpubsub.OpenTopic(conn, topicName, nil)
		default:
			topic, err = pubsub.OpenTopic(ctx, urlStr)
		}
	}
	if err != nil {
		log.Printf("[ERROR] Error connecting to topic: %v\n URL: %s topic: %s\n", err, urlStr, topicName)
	} else {
		log.Printf("[DEBUG] OpenTopic successful %s\n", topicName)
	}

	api.rabbitmqTopicConn = conn
	return topic, err
}

// OpenSubscription opens a subscription to a queue
// Shutdown(ctx) the subscription when you're done with it, and ack() messages when you've processed them
func (api *API) OpenSubscription(ctx context.Context, queue string) (*pubsub.Subscription, error) {
	cnt := 0
	err := errors.New("unknown error")
	urlStr := api.pubSubConfig.QueueURL(queue)
	conn := api.rabbitmqSubscriptionConn
	var subscription *pubsub.Subscription
	for err != nil && cnt <= 5 {
		if cnt > 0 {
			time.Sleep(time.Duration(math.Pow(2.0, float64(cnt))) * time.Second)
		}
		cnt++

		if strings.HasPrefix(urlStr, "awssqs:") {
			subscription, err = pubsub.OpenSubscription(ctx, urlStr)
		} else { // rabbit
			if conn == nil {
				conn, err = amqp.Dial(urlStr)
				if err != nil {
					log.Printf("[INFO] Rabbit dialed try %d error %v\n", cnt, err)
					conn = nil
					continue
				}
			}
			subscription = rabbitpubsub.OpenSubscription(conn, queue, nil)
		}
	}
	if err != nil {
		log.Printf("[ERROR] Error connecting to subscription: %v\n URL: %s queue: %s\n", err, urlStr, queue)
	} else {
		log.Printf("[DEBUG] OpenSubscription successful %s\n", queue)
	}

	api.rabbitmqSubscriptionConn = conn
	return subscription, err
}
