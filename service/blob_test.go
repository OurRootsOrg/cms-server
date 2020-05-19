package service

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestBlobService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	ctx := context.TODO()
	bucket, err := OpenBucket(ctx, "testbucket", "us-east-1", "127.0.0.1:19000", "minioaccess", "miniosecret", true)
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
