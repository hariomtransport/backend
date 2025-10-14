package utils

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	r2Client     *s3.Client
	r2Bucket     string
	r2PublicBase string
	initOnce     sync.Once
)

// initR2 initializes the R2 client once
func initR2() error {
	var initErr error
	initOnce.Do(func() {
		r2Bucket = os.Getenv("R2_BUCKET")
		accountID := os.Getenv("R2_ACCOUNT_ID")
		r2PublicBase = os.Getenv("R2_PUBLIC_URL") // e.g. https://<bucket>.<account_id>.r2.cloudflarestorage.com

		if r2Bucket == "" || accountID == "" || r2PublicBase == "" {
			initErr = fmt.Errorf("missing required R2 environment variables")
			return
		}

		endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           endpoint,
				SigningRegion: "auto",
			}, nil
		})

		cfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion("auto"), // Important for R2
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				os.Getenv("R2_ACCESS_KEY_ID"),
				os.Getenv("R2_SECRET_ACCESS_KEY"),
				"",
			)),
			config.WithEndpointResolverWithOptions(customResolver),
		)
		if err != nil {
			initErr = fmt.Errorf("failed to load R2 config: %v", err)
			return
		}

		r2Client = s3.NewFromConfig(cfg)
	})
	return initErr
}

// UploadToR2 uploads a file (PDF) to R2 and returns its public URL
func UploadToR2(fileBytes []byte, filename string) (string, error) {
	if err := initR2(); err != nil {
		return "", err
	}

	key := filepath.Base(filename)
	_, err := r2Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(r2Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String("application/pdf"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %v", err)
	}

	// Construct public URL
	fileURL := fmt.Sprintf("%s/%s", strings.TrimRight(r2PublicBase, "/"), url.PathEscape(key))
	return fileURL, nil
}

// DeleteFromR2 deletes a file from R2 by URL
func DeleteFromR2(fileURL string) error {
	if err := initR2(); err != nil {
		return err
	}

	u, err := url.Parse(fileURL)
	if err != nil {
		return fmt.Errorf("invalid file URL: %v", err)
	}
	key := filepath.Base(u.Path)

	_, err = r2Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(r2Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete R2 object: %v", err)
	}
	return nil
}
