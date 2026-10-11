package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	"example.com/786-real-estate/backend/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type R2 struct {
	client     *s3.Client
	bucket     string
	publicBase string
}

func NewR2(cfg config.Config) (*R2, error) {
	if cfg.R2AccountID == "" || cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" || cfg.R2Bucket == "" || cfg.R2PublicBaseURL == "" {
		return nil, nil
	}
	base := strings.TrimRight(cfg.R2PublicBaseURL, "/")
	if parsed, err := url.Parse(base); err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Path != "" {
		return nil, fmt.Errorf("invalid R2_PUBLIC_BASE_URL")
	}
	endpoint := "https://" + cfg.R2AccountID + ".r2.cloudflarestorage.com"
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("configure R2 client: %w", err)
	}
	return &R2{
		client: s3.NewFromConfig(awsCfg, func(options *s3.Options) {
			options.BaseEndpoint = aws.String(endpoint)
			options.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		}),
		bucket:     cfg.R2Bucket,
		publicBase: base,
	}, nil
}

func (r *R2) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, error) {
	// Enforce the policy centrally, including asset imports, so no image can
	// bypass conversion by calling the R2 client directly.
	if strings.HasPrefix(contentType, "image/") {
		optimized, err := OptimizeImage(ctx, body)
		if err != nil {
			return "", err
		}
		key = strings.TrimSuffix(key, path.Ext(key)) + "-" + uuid.NewString() + ".webp"
		body, size, contentType = bytes.NewReader(optimized), int64(len(optimized)), "image/webp"
	}
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(r.bucket),
		Key:          aws.String(key),
		Body:         body,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
		ContentLength: func() *int64 {
			if size > 0 {
				return aws.Int64(size)
			}

			return nil
		}(),
	})
	if err != nil {
		return "", fmt.Errorf("upload object: %w", err)
	}
	return r.publicBase + "/" + path.Clean(key), nil
}

func (r *R2) URL(key string) string {
	return r.publicBase + "/" + path.Clean(key)
}

func (r *R2) DeleteURL(ctx context.Context, publicURL string) error {
	if r == nil || !strings.HasPrefix(publicURL, r.publicBase+"/") {
		return nil
	}
	key := strings.TrimPrefix(publicURL, r.publicBase+"/")
	if key == "" || strings.Contains(key, "..") {
		return fmt.Errorf("invalid R2 object URL")
	}
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}
