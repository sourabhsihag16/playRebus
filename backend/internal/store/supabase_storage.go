package store

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// SupabaseStorage handles image storage using Supabase S3-compatible storage
type SupabaseStorage struct {
	s3Client    *s3.S3
	bucketName  string
	publicURL   string
	region      string
}

// NewSupabaseStorage creates a new Supabase storage instance
func NewSupabaseStorage(bucketName, region, accessKey, secretKey, endpoint, publicURL string) (*SupabaseStorage, error) {
	// Create AWS session with custom endpoint for Supabase
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true), // Required for Supabase
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			"",
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &SupabaseStorage{
		s3Client:   s3.New(sess),
		bucketName: bucketName,
		publicURL:  publicURL,
		region:     region,
	}, nil
}

// SaveImage saves image data to Supabase S3 bucket in a date-wise folder
// Path format: {date}/{index}.png
func (s *SupabaseStorage) SaveImage(date string, index int, imageData []byte) error {
	// Create path: date/index.png (e.g., "2025-12-13/0.png")
	key := fmt.Sprintf("%s/%d.png", date, index)

	// Upload to S3
	_, err := s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(key),
		Body:          bytes.NewReader(imageData),
		ContentType:   aws.String("image/png"),
		ContentLength: aws.Int64(int64(len(imageData))),
		ACL:           aws.String("public-read"), // Make images publicly accessible
	})
	if err != nil {
		return fmt.Errorf("failed to upload image to Supabase: %w", err)
	}

	return nil
}

// GetImageURL returns the public URL for an image stored in Supabase
// URL format: {publicURL}/{date}/{index}.png
func (s *SupabaseStorage) GetImageURL(date string, index int) string {
	key := fmt.Sprintf("%s/%d.png", date, index)
	return fmt.Sprintf("%s/%s", s.publicURL, key)
}

// GetImagePath returns the S3 key path for an image (for reference)
func (s *SupabaseStorage) GetImagePath(date string, index int) string {
	return fmt.Sprintf("%s/%d.png", date, index)
}

// GetImage retrieves an image from Supabase S3 (if needed for serving)
func (s *SupabaseStorage) GetImage(date string, index int) ([]byte, error) {
	key := fmt.Sprintf("%s/%d.png", date, index)

	result, err := s.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get image from Supabase: %w", err)
	}
	defer result.Body.Close()

	imageData, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return imageData, nil
}

// ImageExists checks if an image exists in Supabase S3
func (s *SupabaseStorage) ImageExists(date string, index int) (bool, error) {
	key := fmt.Sprintf("%s/%d.png", date, index)

	_, err := s.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a "not found" error
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "NotFound" || aerr.Code() == s3.ErrCodeNoSuchKey {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check image existence: %w", err)
	}

	return true, nil
}

