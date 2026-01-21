package types

import (
	"bytes"
	"context"
	"time"
)

type Wire[T any] interface {
	ConnectCircuitBreaker(CircuitBreaker[T])

	ConnectGenerator(...Generator[T])

	ConnectLogger(...Logger)

	ConnectSensor(...Sensor[T])

	ConnectSurgeProtector(SurgeProtector[T])

	ConnectTransformer(...Transformer[T])

	GetComponentMetadata() ComponentMetadata

	GetGenerators() []Generator[T]

	GetCircuitBreaker() CircuitBreaker[T]

	GetInputChannel() chan T

	GetOutputBuffer() *bytes.Buffer

	GetOutputChannel() chan T

	IsStarted() bool

	Load() *bytes.Buffer

	LoadAsJSONArray() ([]byte, error)

	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})

	SetComponentMetadata(name string, id string)

	SetConcurrencyControl(bufferSize int, maxRoutines int)

	SetEncoder(e Encoder[T])

	SetInputChannel(chan T)

	SetInsulator(retryFunc func(ctx context.Context, elem T, err error) (T, error), threshold int, interval time.Duration)

	SetOutputChannel(chan T)

	Start(context.Context) error

	Restart(ctx context.Context) error

	Submit(ctx context.Context, elem T) error

	Stop() error

	SetTransformerFactory(factory func() Transformer[T])
}
