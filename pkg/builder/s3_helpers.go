package builder

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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

// S3CompatibleStaticConfig configures a static-credential S3-compatible client.
// This is useful for providers like Storj, MinIO, and other S3 gateways.
type S3CompatibleStaticConfig struct {
	Region         string
	Endpoint       string
	AccessKey      string
	SecretKey      string
	SessionToken   string
	ForcePathStyle *bool
	RequireTLS     bool
	HTTPClient     *http.Client
}

// TLSHTTPClientConfig sets TLS-safe defaults for S3-compatible gateways.
type TLSHTTPClientConfig struct {
	MinVersion         uint16
	MaxVersion         uint16
	Timeout            time.Duration
	InsecureSkipVerify bool
}

// NewTLSHTTPClient returns an HTTP client with explicit TLS settings.
func NewTLSHTTPClient(cfg TLSHTTPClientConfig) *http.Client {
	min := cfg.MinVersion
	if min == 0 {
		min = tls.VersionTLS12
	}
	max := cfg.MaxVersion
	if max == 0 {
		max = tls.VersionTLS13
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         min,
				MaxVersion:         max,
				InsecureSkipVerify: cfg.InsecureSkipVerify,
			},
		},
	}
}

// NewS3ClientStaticCompatible builds an S3 client for S3-compatible endpoints.
// Defaults: Region=us-east-1, ForcePathStyle=true.
func NewS3ClientStaticCompatible(ctx context.Context, cfg S3CompatibleStaticConfig) (*s3.Client, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if cfg.AccessKey == "" {
		return nil, fmt.Errorf("access key is required")
	}
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("secret key is required")
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.RequireTLS {
		u, err := url.Parse(cfg.Endpoint)
		if err != nil {
			return nil, fmt.Errorf("invalid endpoint: %w", err)
		}
		if !strings.EqualFold(u.Scheme, "https") {
			return nil, fmt.Errorf("TLS is required; endpoint must be https")
		}
	}
	forcePathStyle := true
	if cfg.ForcePathStyle != nil {
		forcePathStyle = *cfg.ForcePathStyle
	}
	if cfg.HTTPClient == nil {
		return NewS3ClientStatic(
			ctx,
			cfg.Region,
			cfg.AccessKey,
			cfg.SecretKey,
			cfg.SessionToken,
			cfg.Endpoint,
			forcePathStyle,
		)
	}

	// Inline NewS3ClientStatic with HTTP client override.
	loaders := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken),
		),
		config.WithEndpointResolverWithOptions(sharedResolver(cfg.Endpoint)),
		config.WithHTTPClient(cfg.HTTPClient),
	}
	awsCfg, err := config.LoadDefaultConfig(ctx, loaders...)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) { o.UsePathStyle = forcePathStyle }), nil
}

// StorjS3Config configures a Storj S3 Gateway client.
type StorjS3Config struct {
	Endpoint       string
	Region         string
	AccessKey      string
	SecretKey      string
	SessionToken   string
	ForcePathStyle *bool
	RequireTLS     *bool
	HTTPClient     *http.Client
}

// NewS3ClientStorj builds an S3 client for Storj's S3-compatible gateway.
// Defaults: Endpoint=https://gateway.storjshare.io, Region=us-east-1, ForcePathStyle=true.
func NewS3ClientStorj(ctx context.Context, cfg StorjS3Config) (*s3.Client, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://gateway.storjshare.io"
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	requireTLS := true
	if cfg.RequireTLS != nil {
		requireTLS = *cfg.RequireTLS
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = NewTLSHTTPClient(TLSHTTPClientConfig{
			MinVersion: tls.VersionTLS12,
			Timeout:    30 * time.Second,
		})
	}
	return NewS3ClientStaticCompatible(ctx, S3CompatibleStaticConfig{
		Region:         cfg.Region,
		Endpoint:       cfg.Endpoint,
		AccessKey:      cfg.AccessKey,
		SecretKey:      cfg.SecretKey,
		SessionToken:   cfg.SessionToken,
		ForcePathStyle: cfg.ForcePathStyle,
		RequireTLS:     requireTLS,
		HTTPClient:     cfg.HTTPClient,
	})
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
