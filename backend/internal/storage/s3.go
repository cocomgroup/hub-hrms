package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
	prefix     string
}

type StoredDocument struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	URL         string    `json:"url"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

func NewS3Storage(bucketName, prefix string) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		client:     client,
		bucketName: bucketName,
		prefix:     prefix,
	}, nil
}

// Upload stores a document in S3 and returns metadata
func (s *S3Storage) Upload(ctx context.Context, filename string, contentType string, content io.Reader) (*StoredDocument, error) {
	docID := uuid.New().String()
	key := path.Join(s.prefix, docID, filename)

	// Read content into buffer to get size
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, content)
	if err != nil {
		return nil, fmt.Errorf("read content: %w", err)
	}

	// Upload to S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"document-id":   docID,
			"original-name": filename,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("upload to S3: %w", err)
	}

	return &StoredDocument{
		ID:          docID,
		Key:         key,
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		URL:         fmt.Sprintf("s3://%s/%s", s.bucketName, key),
		UploadedAt:  time.Now(),
	}, nil
}

// Download retrieves a document from S3
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, string, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", fmt.Errorf("download from S3: %w", err)
	}

	contentType := ""
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return result.Body, contentType, nil
}

// GetPresignedURL generates a presigned URL for direct download
func (s *S3Storage) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	result, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("generate presigned URL: %w", err)
	}

	return result.URL, nil
}

// Delete removes a document from S3
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err
}

// List returns all documents with a given prefix
func (s *S3Storage) List(ctx context.Context, prefix string) ([]StoredDocument, error) {
	fullPrefix := path.Join(s.prefix, prefix)

	result, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(fullPrefix),
	})
	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err)
	}

	var docs []StoredDocument
	for _, obj := range result.Contents {
		docs = append(docs, StoredDocument{
			Key:        *obj.Key,
			Size:       *obj.Size,
			UploadedAt: *obj.LastModified,
		})
	}

	return docs, nil
}