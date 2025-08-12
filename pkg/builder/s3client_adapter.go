// pkg/builder/s3client_adapter.go
package builder

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	s3ClientAdapter "github.com/joeydtaylor/electrician/pkg/internal/adapter/s3client"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

////////////////////////
// Adapter constructor +
////////////////////////

// NewS3ClientAdapter creates a new S3 client adapter (read + write capable).
func NewS3ClientAdapter[T any](ctx context.Context, options ...types.S3ClientOption[T]) types.S3ClientAdapter[T] {
	return s3ClientAdapter.NewS3ClientAdapter[T](ctx, options...)
}

func S3ClientAdapterWithS3ClientDeps[T any](deps types.S3ClientDeps) types.S3ClientOption[T] {
	return s3ClientAdapter.WithS3ClientDeps[T](deps)
}

func S3ClientAdapterWithWriterConfig[T any](cfg types.S3WriterConfig) types.S3ClientOption[T] {
	return s3ClientAdapter.WithWriterConfig[T](cfg)
}

func S3ClientAdapterWithReaderConfig[T any](cfg types.S3ReaderConfig) types.S3ClientOption[T] {
	return s3ClientAdapter.WithReaderConfig[T](cfg)
}

func S3ClientAdapterWithSensor[T any](sensor ...types.Sensor[T]) types.S3ClientOption[T] {
	return s3ClientAdapter.WithSensor[T](sensor...)
}

func S3ClientAdapterWithLogger[T any](l ...types.Logger) types.S3ClientOption[T] {
	return s3ClientAdapter.WithLogger[T](l...)
}

func S3ClientAdapterWithBatchSettings[T any](maxRecords, maxBytes int, maxAge time.Duration) types.S3ClientOption[T] {
	return s3ClientAdapter.WithBatchSettings[T](maxRecords, maxBytes, maxAge)
}

func S3ClientAdapterWithFormat[T any](format, compression string) types.S3ClientOption[T] {
	return s3ClientAdapter.WithFormat[T](format, compression)
}

func S3ClientAdapterWithSSE[T any](mode, kmsKey string) types.S3ClientOption[T] {
	return s3ClientAdapter.WithSSE[T](mode, kmsKey)
}

func S3ClientAdapterWithReaderListSettings[T any](prefix, startAfter string, pageSize int32, pollEvery time.Duration) types.S3ClientOption[T] {
	return s3ClientAdapter.WithReaderListSettings[T](prefix, startAfter, pageSize, pollEvery)
}

// Inject AWS client + bucket without exposing internal types.
func S3ClientAdapterWithClientAndBucket[T any](cli *s3.Client, bucket string) types.S3ClientOption[T] {
	return s3ClientAdapter.WithS3ClientDeps[T](types.S3ClientDeps{
		Client:         cli,
		Bucket:         bucket,
		ForcePathStyle: true, // default good for LocalStack/MinIO
	})
}

// Writer-only convenience: set prefix template without exposing internals.
func S3ClientAdapterWithWriterPrefixTemplate[T any](prefix string) types.S3ClientOption[T] {
	return s3ClientAdapter.WithWriterConfig[T](types.S3WriterConfig{
		PrefixTemplate: prefix,
	})
}

/////////////////////////////////////////////
// Compliant S3 client constructors (no env)
/////////////////////////////////////////////

// sharedResolver returns an endpoint resolver that maps BOTH S3 and STS to the same override.
func sharedResolver(endpoint string) aws.EndpointResolverWithOptionsFunc {
	return aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
		switch service {
		case s3.ServiceID, sts.ServiceID:
			return aws.Endpoint{URL: endpoint, HostnameImmutable: true}, nil
		default:
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		}
	})
}

// NewS3ClientStatic creates an S3 client using static credentials.
// If endpoint != "", it's used (LocalStack/MinIO). forcePathStyle=true for emulators.
func NewS3ClientStatic(
	ctx context.Context,
	region string,
	accessKey string,
	secretKey string,
	sessionToken string, // "" if none
	endpoint string, // "" for AWS
	forcePathStyle bool,
) (*s3.Client, error) {
	var loaders []func(*config.LoadOptions) error
	if region != "" {
		loaders = append(loaders, config.WithRegion(region))
	}
	loaders = append(loaders, config.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken),
	))
	if endpoint != "" {
		loaders = append(loaders, config.WithEndpointResolverWithOptions(sharedResolver(endpoint)))
	}
	cfg, err := config.LoadDefaultConfig(ctx, loaders...)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = forcePathStyle }), nil
}

// NewS3ClientAssumeRole creates an S3 client by assuming an IAM role via STS.
// sourceCreds: underlying creds to call STS (static keys, SSO, etc.). If nil, default chain.
// externalID optional. duration capped by role MaxSessionDuration.
func NewS3ClientAssumeRole(
	ctx context.Context,
	region string,
	roleARN string,
	sessionName string,
	duration time.Duration,
	externalID string,
	sourceCreds aws.CredentialsProvider, // nil => default provider chain
	endpoint string, // optional S3/STS endpoint override
	forcePathStyle bool,
) (*s3.Client, error) {
	var loaders []func(*config.LoadOptions) error
	if region != "" {
		loaders = append(loaders, config.WithRegion(region))
	}
	if sourceCreds != nil {
		loaders = append(loaders, config.WithCredentialsProvider(sourceCreds))
	}
	if endpoint != "" {
		loaders = append(loaders, config.WithEndpointResolverWithOptions(sharedResolver(endpoint)))
	}
	baseCfg, err := config.LoadDefaultConfig(ctx, loaders...)
	if err != nil {
		return nil, err
	}

	// STS client also uses the same resolver (so it doesn't go to real AWS).
	stsClient := sts.NewFromConfig(baseCfg)

	provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN, func(o *stscreds.AssumeRoleOptions) {
		if sessionName != "" {
			o.RoleSessionName = sessionName
		}
		if duration > 0 {
			o.Duration = duration
		}
		if externalID != "" {
			o.ExternalID = &externalID
		}
	})

	assumed := baseCfg
	assumed.Credentials = aws.NewCredentialsCache(provider)

	return s3.NewFromConfig(assumed, func(o *s3.Options) { o.UsePathStyle = forcePathStyle }), nil
}

// NewS3ClientWebIdentity assumes a role using an OIDC/WebIdentity token file (e.g., EKS IRSA).
func NewS3ClientWebIdentity(
	ctx context.Context,
	region string,
	roleARN string,
	sessionName string,
	tokenFile string,
	duration time.Duration,
	endpoint string, // optional S3/STS endpoint override
	forcePathStyle bool,
) (*s3.Client, error) {
	var loaders []func(*config.LoadOptions) error
	if region != "" {
		loaders = append(loaders, config.WithRegion(region))
	}
	if endpoint != "" {
		loaders = append(loaders, config.WithEndpointResolverWithOptions(sharedResolver(endpoint)))
	}
	cfg, err := config.LoadDefaultConfig(ctx, loaders...)
	if err != nil {
		return nil, err
	}
	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewWebIdentityRoleProvider(
		stsClient,
		roleARN,
		stscreds.IdentityTokenFile(tokenFile),
		func(o *stscreds.WebIdentityRoleOptions) {
			if sessionName != "" {
				o.RoleSessionName = sessionName
			}
			if duration > 0 {
				o.Duration = duration
			}
		},
	)

	assumed := cfg
	assumed.Credentials = aws.NewCredentialsCache(provider)

	return s3.NewFromConfig(assumed, func(o *s3.Options) { o.UsePathStyle = forcePathStyle }), nil
}

// Writer format options (merged; does not clobber other writer config)
func S3ClientAdapterWithWriterFormatOptions[T any](opts map[string]string) types.S3ClientOption[T] {
	// shallow copy to avoid external mutation
	cp := make(map[string]string, len(opts))
	for k, v := range opts {
		cp[k] = v
	}
	return s3ClientAdapter.WithWriterConfig[T](types.S3WriterConfig{
		FormatOptions: cp,
	})
}

// Reader format options (merged; does not clobber other reader config)
func S3ClientAdapterWithReaderFormatOptions[T any](opts map[string]string) types.S3ClientOption[T] {
	cp := make(map[string]string, len(opts))
	for k, v := range opts {
		cp[k] = v
	}
	return s3ClientAdapter.WithReaderConfig[T](types.S3ReaderConfig{
		FormatOptions: cp,
	})
}
