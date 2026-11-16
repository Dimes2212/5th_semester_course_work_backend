package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ErrBucketNotFound = errors.New("storage bucket not found")

type Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

type Client struct {
	s3     *minio.Client
	bucket string
}

func NewClient(config Config) (*Client, error) {
	s3Client, err := minio.New(
		config.Endpoint,
		&minio.Options{
			Creds: credentials.NewStaticV4(
				config.AccessKeyID,
				config.SecretAccessKey,
				"",
			),
			Secure:       config.UseSSL,
			Region:       config.Region,
			BucketLookup: minio.BucketLookupPath,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create S3 client: %w",
			err,
		)
	}

	return &Client{
		s3:     s3Client,
		bucket: config.Bucket,
	}, nil
}

func (c *Client) CheckBucket(ctx context.Context) error {
	exists, err := c.s3.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("check storage bucket: %w", err)
	}
	if !exists {
		return ErrBucketNotFound
	}

	return nil
}

func (c *Client) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := c.s3.PutObject(ctx, c.bucket, objectKey, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("upload object: %w", err)
	}

	return nil
}

func (c *Client) PresignedGetURL(ctx context.Context, objectKey string, expires time.Duration) (string, error) {
	objectURL, err := c.s3.PresignedGetObject(ctx, c.bucket, objectKey, expires, nil)
	if err != nil {
		return "", fmt.Errorf("create presigned object URL: %w", err)
	}

	return objectURL.String(), nil
}

func (c *Client) DeleteObject(ctx context.Context, objectKey string) error {
	err := c.s3.RemoveObject(ctx, c.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}

	return nil
}
