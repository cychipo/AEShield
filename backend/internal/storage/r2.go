package storage

import (
	"context"
	"fmt"
	"io"
	"log"
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
		awsconfig.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	client := s3.NewFromConfig(r2Cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &R2Client{
		client:    client,
		bucket:    cfg.R2BucketName,
		publicURL: cfg.R2PublicURL,
	}, nil
}

// UploadFile upload một file stream lên R2, trả về storage path
func (r *R2Client) UploadFile(ctx context.Context, key string, body io.Reader, contentType string, size int64) error {
	if size < 0 {
		return fmt.Errorf("invalid content length: %d", size)
	}

	input := &s3.PutObjectInput{
		Bucket:        aws.String(r.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}

	log.Printf("[upload.debug] r2 putobject request bucket=%q key=%q contentLength=%d contentType=%q", r.bucket, key, size, contentType)
	_, err := r.client.PutObject(ctx, input)
	if err != nil {
		log.Printf("[upload.debug] r2 putobject error bucket=%q key=%q contentLength=%d err=%v", r.bucket, key, size, err)
		return fmt.Errorf("failed to upload to R2: %w", err)
	}
	log.Printf("[upload.debug] r2 putobject success bucket=%q key=%q contentLength=%d", r.bucket, key, size)

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
