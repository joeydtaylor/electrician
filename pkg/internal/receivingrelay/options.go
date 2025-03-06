// Package receivingrelay offers a collection of functional options that are used to configure instances
// of a ReceivingRelay. These options allow for flexible and dynamic adjustments to the ReceivingRelay's
// behavior and settings during initialization and runtime. The use of options follows the functional
// options pattern, providing a robust way to create configurable objects without an explosion of parasensors
// in constructors or needing to expose the underlying properties directly.

// Each function in this package returns a closure around the option's effect, which can be applied to
// a ReceivingRelay instance. This approach encapsulates the option's logic inside a function that modifies
// the state of a ReceivingRelay, ensuring that all configurations are applied safely and consistently.

// This file includes options for:
// - Setting the network address of the ReceivingRelay.
// - Configuring output conduits to which the ReceivingRelay sends its processed data.
// - Applying TLS configurations to secure the data communication channels.
// - Adjusting the internal buffer size for incoming data streams.
// - Attaching logging mechanisms to provide insight into the ReceivingRelay's operations and health.
// - Setting custom metadata for the ReceivingRelay, aiding in identification and management.

// Utilizing these options enhances the modularity and maintainability of the ReceivingRelay setup,
// enabling developers to customize and extend functionality as required by their specific use cases.
package receivingrelay

import (
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithAddress returns an option for setting the network address of a ReceivingRelay.
// This is crucial for defining where the ReceivingRelay will listen for incoming data.
//
// Parameters:
//   - address: A string specifying the network address.
//
// Returns:
//
//	An option that sets the address when applied to a ReceivingRelay.
func WithAddress[T any](address string) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.SetAddress(address)
	}
}

// WithOutput configures a ReceivingRelay with one or more output conduits.
// These outputs are typically where processed data is sent after being received.
//
// Parameters:
//   - output: A variadic slice of Submitter[T] instances representing the output conduits.
//
// Returns:
//
//	An option that connects the specified outputs to the ReceivingRelay.
func WithOutput[T any](output ...types.Submitter[T]) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.ConnectOutput(output...)
	}
}

// WithTLSConfig sets the TLS configuration for securing the connections of a ReceivingRelay.
// This option is essential for ensuring encrypted communication.
//
// Parameters:
//   - config: A pointer to a types.TLSConfig struct containing TLS settings.
//
// Returns:
//
//	An option that configures TLS settings when applied to a ReceivingRelay.
func WithTLSConfig[T any](config *types.TLSConfig) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.SetTLSConfig(config)
	}
}

// WithBufferSize adjusts the buffer size of the internal data channel of a ReceivingRelay.
// This can impact how much incoming data can be buffered before being processed.
//
// Parameters:
//   - bufferSize: The size of the buffer as a uint32.
//
// Returns:
//
//	An option that sets the buffer size when applied to a ReceivingRelay.
func WithBufferSize[T any](bufferSize uint32) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.SetDataChannel(bufferSize)
	}
}

// WithLogger attaches one or more loggers to a ReceivingRelay.
// These loggers are used for recording operational events and errors.
//
// Parameters:
//   - logger: A variadic slice of Logger interfaces.
//
// Returns:
//
//	An option that connects loggers to the ReceivingRelay.
func WithLogger[T any](logger ...types.Logger) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.ConnectLogger(logger...)
	}
}

// WithComponentMetadata configures a ReceivingRelay component with custom metadata, including a name and an identifier.
// This function provides an option to set these metadata properties, which can be used for identification,
// logging, or other purposes where metadata is needed for a ReceivingRelay. It uses the SetComponentMetadata method
// internally to apply these settings. If the ReceivingRelay's configuration is frozen (indicating that the component
// has started operation and its configuration should no longer be changed), attempting to set metadata
// will result in a panic. This ensures the integrity of component configurations during runtime.
//
// Parameters:
//   - name: The name to set for the ReceivingRelay component, used for identification and logging.
//   - id: The unique identifier to set for the ReceivingRelay component, used for unique identification across systems.
//
// Returns:
//
//	A function conforming to types.Option[types.ReceivingRelay[T]], which when called with a ReceivingRelay component,
//	sets the specified name and id in the component's metadata.
func WithComponentMetadata[T any](name string, id string) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.SetComponentMetadata(name, id)
	}
}
