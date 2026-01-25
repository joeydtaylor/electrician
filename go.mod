module github.com/joeydtaylor/electrician

go 1.24.3

require (
	github.com/andybalholm/brotli v1.1.0
	github.com/aws/aws-sdk-go-v2 v1.38.0
	github.com/aws/aws-sdk-go-v2/config v1.31.0
	github.com/aws/aws-sdk-go-v2/credentials v1.18.4
	github.com/aws/aws-sdk-go-v2/service/s3 v1.87.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.37.0
	github.com/golang/snappy v0.0.4
	github.com/improbable-eng/grpc-web v0.15.0
	github.com/klauspost/compress v1.17.9
	github.com/mjibson/go-dsp v0.0.0-20180508042940-11479a337f12
	github.com/parquet-go/parquet-go v0.25.1
	github.com/pierrec/lz4 v2.6.1+incompatible
	github.com/segmentio/kafka-go v0.4.48
	github.com/shirou/gopsutil v3.21.11+incompatible
	go.uber.org/zap v1.27.0
	golang.org/x/net v0.34.0
	gonum.org/v1/gonum v0.15.0
	google.golang.org/grpc v1.71.0
	google.golang.org/protobuf v1.36.4
	nhooyr.io/websocket v1.8.6
)

replace nhooyr.io/websocket => github.com/coder/websocket v1.8.6

replace github.com/lyft/protoc-gen-validate => github.com/envoyproxy/protoc-gen-validate v0.0.13

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.28.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.33.0 // indirect
	github.com/aws/smithy-go v1.22.5 // indirect
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/frankban/quicktest v1.14.6 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.8.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
)
