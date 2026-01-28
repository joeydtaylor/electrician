package types

import "context"

// Generator defines a function type that continuously produces data items of type T into a system,
// controlled by a context for cancellation and a submission function for emitting data.
type AdapterFunc[T any] func(ctx context.Context, submitFunc func(ctx context.Context, elem T) error)

type Plug[T any] interface {
	AddAdapterFunc(...AdapterFunc[T])
	GetAdapterFuncs() []AdapterFunc[T]
	ConnectSensor(...Sensor[T])
	ConnectLogger(...Logger)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	ConnectAdapter(pc ...Adapter[T])
	GetComponentMetadata() ComponentMetadata

	// SetComponentMetadata sets the metadata for the Forward Relay, such as its name and ID.
	SetComponentMetadata(name string, id string)
	GetConnectors() []Adapter[T]
}
