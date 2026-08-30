package s3

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/net/context"
)

type Client struct {
	s3Client *s3.Client
	bucket   string
}

func NewClient() (*Client, error) {
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		return nil, errors.New("S3_BUCKET environment variable not set")
	}

	cfg := aws.Config{
		Region: os.Getenv("S3_REGION"),
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("S3_ACCESS_KEY"),
			os.Getenv("S3_SECRET_KEY"),
			"",
		),
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(os.Getenv("S3_ENDPOINT_URL"))
	})

	return &Client{
		s3Client: client,
		bucket:   bucket,
	}, nil
}

func (c *Client) Put(ctx context.Context, key string, value []byte) error {
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(value),
	})
	if err != nil {
		return fmt.Errorf("Error putting object: %v", err)
	}
	return nil
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	resp, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("Error getting object: %v", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("Error deleting object: %v", err)
	}
	return nil
}

func (c *Client) List(ctx context.Context, prefix string) ([]string, error) {
	resp, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("Error listing objects: %v", err)
	}
	var keys []string
	for _, obj := range resp.Contents {
		keys = append(keys, *obj.Key)
	}
	return keys, nil
}

func (c *Client) Close() error {
	return nil
}
