package api

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

const internalMinioPrefix = "minio:"
const externalMinioPrefix = "127.0.0.1:"

// OpenBucket opens a blob storage bucket; Close() the bucket when you're done with it
// endpoint, accessKeyID, and secretAccessKey should be empty for AWS
// disableSSL should be true for minio, false for AWS
// set external to true when using this bucket for signing URLs that are used externally, such as in the browser
func (api *API) OpenBucket(ctx context.Context, external bool) (*blob.Bucket, error) {
	var creds *credentials.Credentials
	if api.blobStoreConfig.accessKey != "" && api.blobStoreConfig.secretKey != "" {
		creds = credentials.NewStaticCredentials(api.blobStoreConfig.accessKey, api.blobStoreConfig.secretKey, "")
	}
	endpoint := api.blobStoreConfig.endpoint
	if external && strings.HasPrefix(endpoint, internalMinioPrefix) {
		endpoint = externalMinioPrefix + endpoint[len(internalMinioPrefix):]
	}
	config := aws.Config{
		Region:           aws.String(api.blobStoreConfig.region),
		Endpoint:         aws.String(endpoint),
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
