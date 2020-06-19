package api

import (
	"context"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"gocloud.dev/pubsub"
)

func TestPubSubService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	ctx := context.TODO()
	content := "Hello world!"

	ap, err := NewAPI()
	if err != nil {
		log.Fatalf("Error calling NewAPI: %v", err)
	}
	defer ap.Close()
	ap = ap.QueueConfig("test", "amqp://guest:guest@localhost:35672/")

	// publish to queue
	topic, err := ap.OpenTopic(ctx, "test")
	assert.NoError(t, err)
	defer topic.Shutdown(ctx)

	err = topic.Send(ctx, &pubsub.Message{
		Body: []byte(content),
	})
	assert.NoError(t, err)

	// read from queue
	sub, err := ap.OpenSubscription(ctx, "test")
	assert.NoError(t, err)
	defer sub.Shutdown(ctx)

	msg, err := sub.Receive(ctx)
	assert.NoError(t, err)
	msg.Ack()
	assert.Equal(t, content, string(msg.Body))
}
