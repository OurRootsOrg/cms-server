package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gocloud.dev/pubsub"
	"testing"
)

func TestPubSubService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	ctx := context.TODO()
	urlStr := "rabbit://test"
	content := "Hello world!"

	// publish to queue
	topic, err := OpenTopic(ctx, urlStr)
	assert.NoError(t, err)
	defer topic.Shutdown(ctx)

	err = topic.Send(ctx, &pubsub.Message{
		Body: []byte(content),
	})
	assert.NoError(t, err)

	// read from queue
	sub, err := OpenSubscription(ctx, urlStr)
	assert.NoError(t, err)
	defer sub.Shutdown(ctx)

	msg, err := sub.Receive(ctx)
	assert.NoError(t, err)
	msg.Ack()
	assert.Equal(t, content, string(msg.Body))
}

