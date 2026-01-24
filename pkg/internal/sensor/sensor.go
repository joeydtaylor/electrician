package sensor

import (
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// Sensor provides callback hooks for component telemetry.
type Sensor[T any] struct {
	componentMetadata types.ComponentMetadata
	metadataLock      sync.Mutex

	OnStart                               []func(types.ComponentMetadata)
	OnInsulatorAttempt                    []func(types.ComponentMetadata, T, T, error, error, int, int, time.Duration)
	OnInsulatorSuccess                    []func(types.ComponentMetadata, T, T, error, error, int, int, time.Duration)
	OnInsulatorFailure                    []func(types.ComponentMetadata, T, T, error, error, int, int, time.Duration)
	OnRestart                             []func(types.ComponentMetadata)
	OnSubmit                              []func(types.ComponentMetadata, T)
	OnElementProcessed                    []func(types.ComponentMetadata, T)
	OnError                               []func(types.ComponentMetadata, error, T)
	OnCancel                              []func(types.ComponentMetadata, T)
	OnComplete                            []func(types.ComponentMetadata)
	OnTerminate                           []func(types.ComponentMetadata)
	OnHTTPClientRequestStart              []func(types.ComponentMetadata)
	OnHTTPClientResponseReceived          []func(types.ComponentMetadata)
	OnHTTPClientError                     []func(types.ComponentMetadata, error)
	OnHTTPClientRequestComplete           []func(types.ComponentMetadata)
	OnSurgeProtectorTrip                  []func(types.ComponentMetadata)
	OnSurgeProtectorReset                 []func(types.ComponentMetadata)
	OnSurgeProtectorBackupFailure         []func(types.ComponentMetadata, error)
	OnSurgeProtectorRateLimitExceeded     []func(types.ComponentMetadata, T)
	OnResisterDequeued                    []func(types.ComponentMetadata, T)
	OnResisterQueued                      []func(types.ComponentMetadata, T)
	OnResisterRequeued                    []func(types.ComponentMetadata, T)
	OnResisterEmpty                       []func(types.ComponentMetadata)
	OnSurgeProtectorReleaseToken          []func(types.ComponentMetadata)
	OnSurgeProtectorConnectResister       []func(types.ComponentMetadata, types.ComponentMetadata)
	OnSurgeProtectorDetachedBackups       []func(types.ComponentMetadata, types.ComponentMetadata)
	OnSurgeProtectorBackupWireSubmit      []func(types.ComponentMetadata, T)
	OnSurgeProtectorSubmit                []func(types.ComponentMetadata, T)
	OnSurgeProtectorDrop                  []func(types.ComponentMetadata, T)
	OnCircuitBreakerTrip                  []func(types.ComponentMetadata, int64, int64)
	OnCircuitBreakerReset                 []func(types.ComponentMetadata, int64)
	OnCircuitBreakerNeutralWireSubmission []func(types.ComponentMetadata, T)
	OnCircuitBreakerRecordError           []func(types.ComponentMetadata, int64)
	OnCircuitBreakerAllow                 []func(types.ComponentMetadata)
	OnCircuitBreakerDrop                  []func(types.ComponentMetadata, T)

	OnS3WriterStart       []func(types.ComponentMetadata, string, string, string)
	OnS3WriterStop        []func(types.ComponentMetadata)
	OnS3KeyRendered       []func(types.ComponentMetadata, string)
	OnS3PutAttempt        []func(types.ComponentMetadata, string, string, int, string, string)
	OnS3PutSuccess        []func(types.ComponentMetadata, string, string, int, time.Duration)
	OnS3PutError          []func(types.ComponentMetadata, string, string, int, error)
	OnS3ParquetRollFlush  []func(types.ComponentMetadata, int, int, string)
	OnS3ReaderListStart   []func(types.ComponentMetadata, string, string)
	OnS3ReaderListPage    []func(types.ComponentMetadata, int, bool)
	OnS3ReaderObject      []func(types.ComponentMetadata, string, int64)
	OnS3ReaderDecode      []func(types.ComponentMetadata, string, int, string)
	OnS3ReaderSpillToDisk []func(types.ComponentMetadata, int64, int64)
	OnS3ReaderComplete    []func(types.ComponentMetadata, int, int)
	OnS3BillingSample     []func(types.ComponentMetadata, string, int64, int64, string)

	OnKafkaWriterStart []func(types.ComponentMetadata, string, string)
	OnKafkaWriterStop  []func(types.ComponentMetadata)

	OnKafkaProduceAttempt []func(types.ComponentMetadata, string, int, int, int)
	OnKafkaProduceSuccess []func(types.ComponentMetadata, string, int, int64, time.Duration)
	OnKafkaProduceError   []func(types.ComponentMetadata, string, int, error)
	OnKafkaBatchFlush     []func(types.ComponentMetadata, string, int, int, string)

	OnKafkaKeyRendered     []func(types.ComponentMetadata, []byte)
	OnKafkaHeadersRendered []func(types.ComponentMetadata, []struct{ Key, Value string })

	OnKafkaConsumerStart     []func(types.ComponentMetadata, string, []string, string)
	OnKafkaConsumerStop      []func(types.ComponentMetadata)
	OnKafkaPartitionAssigned []func(types.ComponentMetadata, string, int, int64, int64)
	OnKafkaPartitionRevoked  []func(types.ComponentMetadata, string, int)
	OnKafkaMessage           []func(types.ComponentMetadata, string, int, int64, int, int)
	OnKafkaDecode            []func(types.ComponentMetadata, string, int, string)

	OnKafkaCommitSuccess []func(types.ComponentMetadata, string, map[string]int64)
	OnKafkaCommitError   []func(types.ComponentMetadata, string, error)

	OnKafkaDLQProduceAttempt []func(types.ComponentMetadata, string, int, int, int)
	OnKafkaDLQProduceSuccess []func(types.ComponentMetadata, string, int, int64, time.Duration)
	OnKafkaDLQProduceError   []func(types.ComponentMetadata, string, int, error)

	OnKafkaBillingSample []func(types.ComponentMetadata, string, int64, int64)

	callbackLock sync.Mutex
	loggers      []types.Logger
	loggersLock  sync.Mutex
	meters       []types.Meter[T]
	metersLock   sync.Mutex
}

// NewSensor constructs a Sensor with optional configuration.
func NewSensor[T any](options ...types.Option[types.Sensor[T]]) types.Sensor[T] {
	s := &Sensor[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "SENSOR",
		},
	}

	for _, opt := range s.decorateCallbacks(options...) {
		if opt == nil {
			continue
		}
		opt(s)
	}

	return s
}
