package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/ourrootsorg/cms-server/utils"

	"github.com/stretchr/testify/assert"
)

func TestBlobService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	ctx := utils.AddSocietyIDToContext(context.TODO(), 1)

	ap, err := NewAPI()
	if err != nil {
		log.Fatalf("Error calling NewAPI: %v", err)
	}
	defer ap.Close()
	ap = ap.
		BlobStoreConfig("us-east-1", "127.0.0.1:19000", "minioaccess", "miniosecret", "testbucket", true)

	bucket, err := ap.OpenBucket(ctx, false)
	assert.NoError(t, err)
	defer bucket.Close()

	// write an object
	content := "Hello, World!"
	w, err := bucket.NewWriter(ctx, "test.txt", nil)
	assert.NoError(t, err)
	_, err = fmt.Fprint(w, content)
	assert.NoError(t, err)
	err = w.Close()
	assert.NoError(t, err)

	// read an object
	r, err := bucket.NewReader(ctx, "test.txt", nil)
	assert.NoError(t, err)
	defer r.Close()
	data, err := ioutil.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, content, string(data))
}
