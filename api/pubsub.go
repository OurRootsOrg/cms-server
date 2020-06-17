package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
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
	urlStr := getPubSubURLStr(api.pubSubConfig.protocol, api.pubSubConfig.region, api.pubSubConfig.host, topicName)
	conn := api.rabbitmqTopicConn
	var topic *pubsub.Topic
	for err != nil && cnt <= 5 {
		if cnt > 0 {
			time.Sleep(time.Duration(math.Pow(2.0, float64(cnt))) * time.Second)
		}
		cnt++

		if api.pubSubConfig.protocol == "awssqs" {
			topic, err = pubsub.OpenTopic(ctx, urlStr)
		} else { // rabbit
			if conn == nil {
				conn, err = amqp.Dial(urlStr)
				if err != nil {
					log.Printf("[INFO] Rabbit dialed try %d error %v\n", cnt, err)
					conn = nil
					continue
				}
			}
			topic = rabbitpubsub.OpenTopic(conn, topicName, nil)
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
	urlStr := getPubSubURLStr(api.pubSubConfig.protocol, api.pubSubConfig.region, api.pubSubConfig.host, queue)
	conn := api.rabbitmqSubscriptionConn
	var subscription *pubsub.Subscription
	for err != nil && cnt <= 5 {
		if cnt > 0 {
			time.Sleep(time.Duration(math.Pow(2.0, float64(cnt))) * time.Second)
		}
		cnt++

		if api.pubSubConfig.protocol == "awssqs" {
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

func getPubSubURLStr(protocol, region, host, target string) string {
	if protocol == "awssqs" {
		return fmt.Sprintf("awssqs://%s/%s?region=%s", host, target, region)
	}
	return fmt.Sprintf("amqp://%s/", host)
}
