package storage

import (
	"context"
	"net/url"
	"testing"

	"example.com/786-real-estate/backend/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestR2UsesAccountEndpointAndBucketPath(t *testing.T) {
	r, err := NewR2(config.Config{R2AccountID: "test-account", R2AccessKeyID: "test-key", R2SecretAccessKey: "test-secret", R2Bucket: "test-bucket", R2PublicBaseURL: "https://assets.example.test"})
	if err != nil {
		t.Fatal(err)
	}
	request, err := s3.NewPresignClient(r.client).PresignPutObject(context.Background(), &s3.PutObjectInput{Bucket: aws.String(r.bucket), Key: aws.String("documents/test.pdf")})
	if err != nil {
		t.Fatal(err)
	}
	target, err := url.Parse(request.URL)
	if err != nil {
		t.Fatal("invalid upload URL")
	}
	if target.Host != "test-account.r2.cloudflarestorage.com" || target.Path != "/test-bucket/documents/test.pdf" {
		t.Fatalf("wrong storage addressing: %s%s", target.Host, target.Path)
	}
}
