package api

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

// OpenBucket opens a blob storage bucket; Close() the bucket when you're done with it
// endpoint, accessKeyID, and secretAccessKey should be empty for AWS
// disableSSL should be true for minio, false for AWS
func (api *API) OpenBucket(ctx context.Context) (*blob.Bucket, error) {
	var creds *credentials.Credentials
	if api.blobStoreConfig.accessKey != "" && api.blobStoreConfig.secretKey != "" {
		creds = credentials.NewStaticCredentials(api.blobStoreConfig.accessKey, api.blobStoreConfig.secretKey, "")
	}
	config := aws.Config{
		Region:           aws.String(api.blobStoreConfig.region),
		Endpoint:         aws.String(api.blobStoreConfig.endpoint),
		DisableSSL:       aws.Bool(api.blobStoreConfig.disableSSL),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      creds,
	}
	sess, err := session.NewSession(&config)
	if err != nil {
		return nil, err
	}

	return s3blob.OpenBucket(ctx, sess, api.blobStoreConfig.bucket, nil)
}
