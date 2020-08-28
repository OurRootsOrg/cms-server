package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting/unit"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ourrootsorg/cms-server/api"
	"github.com/stretchr/testify/assert"
)

func dlLoggingSvc(data []byte) (*s3.S3, *[]string, *[]string) {
	var m sync.Mutex
	names := []string{}
	ranges := []string{}

	svc := s3.New(unit.Session)
	svc.Handlers.Send.Clear()
	svc.Handlers.Send.PushBack(func(r *request.Request) {
		m.Lock()
		defer m.Unlock()
		names = append(names, r.Operation.Name)

		switch op := r.Params.(type) {
		case *s3.GetObjectInput:
			ranges = append(ranges, *op.Range)
			rerng := regexp.MustCompile(`bytes=(\d+)-(\d+)`)
			log.Printf("Range: %s", r.HTTPRequest.Header.Get("Range"))
			rng := rerng.FindStringSubmatch(r.HTTPRequest.Header.Get("Range"))
			log.Printf("rng: %#v", rng)
			start, _ := strconv.ParseInt(rng[1], 10, 64)
			fin, _ := strconv.ParseInt(rng[2], 10, 64)
			fin++

			if fin > int64(len(data)) {
				fin = int64(len(data))
			}

			bodyBytes := data[start:fin]
			r.HTTPResponse = &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(bodyBytes)),
				Header:     http.Header{},
			}
			r.HTTPResponse.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d",
				start, fin-1, len(data)))
			r.HTTPResponse.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))
		case *s3.HeadObjectInput:
			r.HTTPResponse = &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
				Header:     http.Header{},
			}
			r.HTTPResponse.Header.Set("Content-Length", fmt.Sprintf("%d", len(data)))
		}
	})

	return svc, &names, &ranges
}

func TestS3ReaderAt(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping tests in short mode")
	}
	ctx := context.TODO()

	ap, err := api.NewAPI()
	if err != nil {
		log.Fatalf("Error calling NewAPI: %v", err)
	}
	defer ap.Close()
	ap = ap.
		BlobStoreConfig("us-east-1", "127.0.0.1:19000", "minioaccess", "miniosecret", "testbucket", true)

	bucket, err := ap.OpenBucket(ctx)
	assert.NoError(t, err)
	defer bucket.Close()

	zipBytes, err := ioutil.ReadFile("testdata/test.zip")
	assert.NoError(t, err)
	t.Logf("len(zipBytes): %d\n", len(zipBytes))

	// write an object
	w, err := bucket.NewWriter(ctx, "test.zip", nil)
	assert.NoError(t, err)
	io.Copy(w, bytes.NewBuffer(zipBytes))
	assert.NoError(t, err)
	err = w.Close()
	assert.NoError(t, err)

	ra, err := NewBucketReaderAt(ctx, bucket, "test.zip")
	assert.NoError(t, err)
	t.Logf("ra.Size: %d\n", ra.Size)
	zr, err := zip.NewReader(ra, ra.Size)
	assert.NoError(t, err)
	for _, f := range zr.File {
		log.Printf("File: %s, size: %d\n", f.Name, f.UncompressedSize)
		rc, err := f.Open()
		if err != nil {
			assert.NoError(t, err)
			return
		}
		t.Logf("rc: %v", rc)
		defer rc.Close()
		b, err := ioutil.ReadAll(rc)
		if err != nil && err != io.EOF {
			assert.NoError(t, err)
		}
		assert.Equal(t, int(f.UncompressedSize), len(b))
	}
}
