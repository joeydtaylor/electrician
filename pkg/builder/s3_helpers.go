package builder

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// LocalstackS3AssumeRoleConfig sets up defaults for LocalStack assume-role clients.
type LocalstackS3AssumeRoleConfig struct {
	RoleARN      string
	SessionName  string
	Region       string
	Duration     time.Duration
	ExternalID   string
	Endpoint     string
	AccessKey    string
	SecretKey    string
	SessionToken string
}

// NewS3ClientAssumeRoleLocalstack builds an assume-role S3 client with LocalStack defaults.
func NewS3ClientAssumeRoleLocalstack(ctx context.Context, cfg LocalstackS3AssumeRoleConfig) (*s3.Client, error) {
	if cfg.RoleARN == "" {
		return nil, fmt.Errorf("role ARN is required")
	}
	if cfg.SessionName == "" {
		cfg.SessionName = "electrician"
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.Duration == 0 {
		cfg.Duration = 15 * time.Minute
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://localhost:4566"
	}
	if cfg.AccessKey == "" {
		cfg.AccessKey = "test"
	}
	if cfg.SecretKey == "" {
		cfg.SecretKey = "test"
	}

	creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken))
	return NewS3ClientAssumeRole(
		ctx,
		cfg.Region,
		cfg.RoleARN,
		cfg.SessionName,
		cfg.Duration,
		cfg.ExternalID,
		creds,
		cfg.Endpoint,
		true,
	)
}

// S3ListKeys returns object keys for a bucket/prefix, optionally filtered by suffix.
func S3ListKeys(ctx context.Context, cli *s3.Client, bucket, prefix string, suffixes ...string) ([]string, error) {
	if cli == nil {
		return nil, fmt.Errorf("s3 client is required")
	}
	if bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	var keys []string
	var cont *string

	for {
		out, err := cli.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: cont,
			MaxKeys:           aws.Int32(1000),
		})
		if err != nil {
			return nil, err
		}
		for _, o := range out.Contents {
			k := aws.ToString(o.Key)
			if len(suffixes) == 0 || hasSuffixFold(k, suffixes) {
				keys = append(keys, k)
			}
		}
		if aws.ToBool(out.IsTruncated) {
			cont = out.NextContinuationToken
			continue
		}
		break
	}

	return keys, nil
}

func hasSuffixFold(key string, suffixes []string) bool {
	lower := strings.ToLower(key)
	for _, s := range suffixes {
		if strings.HasSuffix(lower, strings.ToLower(s)) {
			return true
		}
	}
	return false
}
