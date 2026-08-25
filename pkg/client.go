// Client-level operations

package pkg

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Config holds the configuration for the R2 client. This is used to authenticate and connect to the
// R2 API. The profile is the name of the profile in the ~/.r2 configuration file. The account ID is
// the ID of the R2 account. The access key ID and secret access key are the credentials for the
// account.
type Config struct {
	Profile         string
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
}

// R2Client is a wrapper around the S3 client that provides methods for interacting with R2. This
// allows us to add methods to the client. The S3 client is embedded in the R2Client struct so that
// we can use the existing methods of the S3 client without having to re-implement them.
type R2Client struct {
	s3.Client
}

// s3Client returns a new S3 client for the given profile. The client is configured with the R2
// endpoint and credentials for the given profile. This is used to create the R2Client and
// R2PresignClient structs, which are used for all R2 operations.
func s3Client(c Config) *s3.Client {
	// R2 requires a dummy region - using "auto" as it's Cloudflare's convention
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion("auto"),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, "")),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create S3 client with custom R2 endpoint
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", c.AccountID))
		o.UsePathStyle = true
	})
}

// Client returns a new R2 client struct so we can add methods to it. The client is configured with
// the R2 endpoint and credentials for the given profile.
func Client(c Config) R2Client {
	return R2Client{*s3Client(c)}
}

// R2PresignClient is a wrapper around the S3 presign client that provides methods for interacting
// with R2. This allows us to add methods to the client. The S3 presign client is embedded in the
// R2PresignClient struct so that we can use the existing methods of the S3 presign client without
// having to re-implement them.
type R2PresignClient struct {
	s3.PresignClient
}

// PresignClient returns a new R2 presign client struct so we can add methods to it. The client is
// configured with the R2 endpoint and credentials for the given profile. The presign client is
// used for generating presigned URLs.
func PresignClient(c Config) R2PresignClient {
	s3c := s3Client(c)
	return R2PresignClient{*s3.NewPresignClient(
		s3c,
		s3.WithPresignClientFromClientOptions(func(o *s3.Options) {
			// The CLI returns presigned GETs as URLs only. The SDK's default response
			// checksum validation adds x-amz-checksum-mode to the signed headers,
			// which URL-only consumers cannot reproduce.
			o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
		}),
	)}
}

// PrintBuckets prints the creation date and name of each bucket in the R2 account.
func (c *R2Client) PrintBuckets() {
	// Get buckets
	listBucketsOutput, err := c.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}

	// Print creation date and name of each bucket
	for _, object := range listBucketsOutput.Buckets {
		fmt.Println(object.CreationDate.Format("2006-01-02 15:04:05"), *object.Name)
	}
}

// MakeBucket creates a new R2 bucket with the given name. The bucket is created in the account
// associated with the R2 client. The bucket name must be unique across all existing bucket names in
// the account.
func (c *R2Client) MakeBucket(name string) {
	_, err := c.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket:                    aws.String(name),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{},
	})
	if err != nil {
		log.Fatalf("Error creating bucket %s: %v\n", name, err)
	}
}

// RemoveBucket removes the bucket with the given name from the R2 account. The bucket must be empty
// before it can be removed.
func (c *R2Client) RemoveBucket(bucket string) {
	_, err := c.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucket)})
	if err != nil {
		log.Fatalf("Couldn't delete bucket %s: %v\n", bucket, err)
	}
}
