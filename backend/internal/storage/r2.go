package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aeshield/backend/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Client struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewR2Client(cfg *config.Config) (*R2Client, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID)

	r2Cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, ""),
		),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	client := s3.NewFromConfig(r2Cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &R2Client{
		client:    client,
		bucket:    cfg.R2BucketName,
		publicURL: cfg.R2PublicURL,
	}, nil
}

// UploadFile upload một file stream lên R2, trả về storage path
func (r *R2Client) UploadFile(ctx context.Context, key string, body io.Reader, contentType string, size int64) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}
	if size > 0 {
		input.ContentLength = aws.Int64(size)
	}

	_, err := r.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload to R2: %w", err)
	}

	return nil
}

// DeleteFile xóa file khỏi R2
func (r *R2Client) DeleteFile(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from R2: %w", err)
	}
	return nil
}

// GeneratePresignedURL tạo presigned URL để download file (mặc định 1 giờ)
func (r *R2Client) GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(r.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

// PublicURL trả về public URL của file (chỉ dùng khi bucket enable public access)
func (r *R2Client) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s", r.publicURL, key)
}
