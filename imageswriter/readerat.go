package main

import (
	"context"
	"io"
	"log"

	"gocloud.dev/blob"
)

// BucketReaderAt implements an io.ReaderAt on top of the AWS S3 API
type BucketReaderAt struct {
	bucket *blob.Bucket
	Key    string
	Size   int64
}

// NewBucketReaderAt creates a BucketReaderAt
func NewBucketReaderAt(ctx context.Context, bucket *blob.Bucket, key string) (*BucketReaderAt, error) {
	var r BucketReaderAt
	r.bucket = bucket
	r.Key = key
	attr, err := bucket.Attributes(ctx, key)
	if err != nil {
		return nil, err
	}
	r.Size = attr.Size
	return &r, nil
}

// ReadAt implements io.ReaderAt.ReadAt by reading from an S3 bucket
func (r *BucketReaderAt) ReadAt(p []byte, off int64) (int, error) {
	cnt := len(p)
	// log.Printf("ReadAt(p[%d], %d)", cnt, off)
	rr, err := r.bucket.NewRangeReader(context.TODO(), r.Key, off, int64(cnt), nil)
	if err != nil {
		log.Printf("NewRangeReader(ctx, %s, %d, %d) returned err %#v", r.Key, off, cnt, err)
		return 0, err
	}
	defer rr.Close()
	n := 0
	for n <= cnt && err == nil {
		var nn int
		nn, err = rr.Read(p[n:])
		n += nn
		// log.Printf("nn %d, err %#v, n %d", nn, err, n)
	}
	if n == cnt && err == io.EOF {
		err = nil
	}
	return n, err
}
