package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

// AWSV2Uploader implements Uploader using aws-sdk-go-v2 S3 client.
type AWSV2Uploader struct {
	client        *awss3.Client
	region        string
	endpoint      string
	usePathStyle  bool
	publicBaseURL string
}

// NewAWSV2Uploader constructs an aws-sdk-go-v2 S3 client and wraps it to conform to Uploader.
// - region: AWS region (e.g., "us-east-1"). Required for AWS; for S3-compatible endpoints, supply any region string.
// - endpoint: Optional custom endpoint (e.g., "https://s3.us-west-002.backblazeb2.com"). Empty for AWS default.
// - usePathStyle: Whether to use path-style URLs with custom endpoints.
// - publicBaseURL: Optional absolute base URL used to form public URLs. If empty, a URL will be inferred.
func NewAWSV2Uploader(ctx context.Context, region, endpoint string, usePathStyle bool, publicBaseURL string) (*AWSV2Uploader, error) {
	if region == "" {
		region = "us-east-1"
	}

	// Load default AWS config; region is needed even for custom endpoints
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	// Build S3 client with optional endpoint override and path-style setting
	client := awss3.NewFromConfig(cfg, func(o *awss3.Options) {
		o.UsePathStyle = usePathStyle
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	return &AWSV2Uploader{
		client:        client,
		region:        region,
		endpoint:      strings.TrimRight(endpoint, "/"),
		usePathStyle:  usePathStyle,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}, nil
}

func (u *AWSV2Uploader) PutObject(ctx context.Context, bucket, key string, body io.Reader, contentType string) (string, error) {
	input := &awss3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   body,
	}
	if contentType != "" {
		input.ContentType = &contentType
	}

	if _, err := u.client.PutObject(ctx, input); err != nil {
		return "", err
	}

	// Build public URL
	urlStr := u.objectURL(bucket, key)
	return urlStr, nil
}

func (u *AWSV2Uploader) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := u.client.DeleteObject(ctx, &awss3.DeleteObjectInput{Bucket: &bucket, Key: &key})
	return err
}

func (u *AWSV2Uploader) objectURL(bucket, key string) string {
	if u.publicBaseURL != "" {
		return u.publicBaseURL + "/" + escapePath(key)
	}
	// If custom endpoint provided, prefer path-style when requested
	if u.endpoint != "" {
		if u.usePathStyle {
			return u.endpoint + "/" + bucket + "/" + escapePath(key)
		}
		// Virtual-hosted-style for custom endpoint
		// e.g., https://bucket.endpoint/key
		return "https://" + bucket + "." + trimScheme(u.endpoint) + "/" + escapePath(key)
	}
	// Default AWS URL
	// https://{bucket}.s3.{region}.amazonaws.com/{key}
	host := fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, u.region)
	return host + "/" + escapePath(key)
}

func trimScheme(raw string) string {
	if strings.HasPrefix(raw, "https://") {
		return strings.TrimPrefix(raw, "https://")
	}
	if strings.HasPrefix(raw, "http://") {
		return strings.TrimPrefix(raw, "http://")
	}
	return raw
}

func escapePath(p string) string {
	// URL-escape only path segments, preserving slashes
	segments := strings.Split(p, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return strings.Join(segments, "/")
}
