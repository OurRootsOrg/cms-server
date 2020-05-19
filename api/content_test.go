package api_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/ourrootsorg/cms-server/api"
	"github.com/stretchr/testify/assert"
)

func TestContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	ctx := context.TODO()
	testApi, err := api.NewAPI()
	assert.NoError(t, err)
	testApi.BlobStoreConfig("us-east-1", "127.0.0.1:19000",
		"minioaccess", "miniosecret", "testbucket", true)
	content := "Hello,World"

	// make the request
	contentRequest, errs := testApi.PostContentRequest(ctx, api.ContentRequest{"text/csv"})
	assert.Nil(t, errs)

	// post the content
	client := &http.Client{}
	req, err := http.NewRequest("PUT", contentRequest.PutURL, strings.NewReader(content))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "text/csv")
	res, err := client.Do(req)
	assert.Equal(t, 200, res.StatusCode)

	// read the content
	result, err := testApi.GetContent(ctx, contentRequest.Key)
	assert.Nil(t, err)
	assert.Equal(t, content, string(result))
}
