package types

import (
	"context"
)

type Adapter[T any] interface {
	Serve(ctx context.Context, submitFunc func(ctx context.Context, elem T) error) error
	ConnectSensor(...Sensor[T])
	ConnectLogger(...Logger)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	GetComponentMetadata() ComponentMetadata

	// SetComponentMetadata sets the metadata for the Forward Relay, such as its name and ID.
	SetComponentMetadata(name string, id string)
}
