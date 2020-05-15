package service

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

// endpoint, accessKeyID, and secretAccessKey should be empty for AWS
// disableSSL should be true for minio, false for AWS
// Close() the bucket when you're done with it
func OpenBucket(ctx context.Context, bucketName, region, endpoint, accessKeyID, secretAccessKey string, disableSSL bool) (*blob.Bucket, error) {
	var creds *credentials.Credentials
	if accessKeyID != "" && secretAccessKey != "" {
		creds = credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
	}
	config := aws.Config{
		Region: aws.String(region),
		Endpoint: aws.String(endpoint),
		DisableSSL: aws.Bool(disableSSL),
		S3ForcePathStyle: aws.Bool(true),
		Credentials: creds,
	}
	sess, err := session.NewSession(&config)
	if err != nil {
		return nil, err
	}

	return s3blob.OpenBucket(ctx, sess, bucketName, nil)
}
