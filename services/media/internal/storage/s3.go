package storage

import (
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/yourorg/b25/services/media/internal/config"
)

// S3Storage implements Storage interface using AWS S3 (or compatible services like MinIO)
type S3Storage struct {
	client *s3.S3
	bucket string
	config config.S3Config
}

// NewS3Storage creates a new S3 storage backend
func NewS3Storage(cfg config.S3Config) (*S3Storage, error) {
	awsConfig := &aws.Config{
		Region:           aws.String(cfg.Region),
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		S3ForcePathStyle: aws.Bool(cfg.ForcePathStyle),
	}

	if cfg.Endpoint != "" {
		awsConfig.Endpoint = aws.String(cfg.Endpoint)
	}

	if !cfg.UseSSL {
		awsConfig.DisableSSL = aws.Bool(true)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	client := s3.New(sess)

	// Check if bucket exists, create if it doesn't
	_, err = client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})

	if err != nil {
		// Try to create bucket
		_, err = client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(cfg.Bucket),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
		config: cfg,
	}, nil
}

// Upload uploads a file to S3
func (s *S3Storage) Upload(path string, reader io.Reader, contentType string) (string, error) {
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(path),
		Body:        aws.ReadSeekCloser(reader),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return path, nil
}

// Download downloads a file from S3
func (s *S3Storage) Download(path string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	return result.Body, nil
}

// Delete removes a file from S3
func (s *S3Storage) Delete(path string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// Exists checks if a file exists in S3
func (s *S3Storage) Exists(path string) (bool, error) {
	_, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		return false, nil
	}

	return true, nil
}

// GetURL returns the public URL for a file
func (s *S3Storage) GetURL(path string) string {
	if s.config.Endpoint != "" {
		// For MinIO or custom endpoints
		protocol := "http"
		if s.config.UseSSL {
			protocol = "https"
		}
		return fmt.Sprintf("%s://%s/%s/%s", protocol, s.config.Endpoint, s.bucket, path)
	}

	// For AWS S3
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.config.Region, path)
}

// GetSignedURL returns a signed URL for temporary access
func (s *S3Storage) GetSignedURL(path string, expiry int) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	url, err := req.Presign(time.Duration(expiry) * time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// Copy copies a file within S3
func (s *S3Storage) Copy(sourcePath, destPath string) error {
	_, err := s.client.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(fmt.Sprintf("%s/%s", s.bucket, sourcePath)),
		Key:        aws.String(destPath),
	})

	if err != nil {
		return fmt.Errorf("failed to copy in S3: %w", err)
	}

	return nil
}

// GetSize returns the size of a file
func (s *S3Storage) GetSize(path string) (int64, error) {
	result, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	return *result.ContentLength, nil
}
