// Package forwardrelay provides options for configuring the ForwardRelay component, enabling users to customize its behavior and functionality.
//
// The options.go file defines various functions, each representing a specific configuration option that can be applied to a ForwardRelay instance.
// These options allow users to set the network address, input conduit, TLS configuration, logger, performance options, and component metadata of a ForwardRelay.
// By leveraging these options, users can achieve fine-grained control over the operation and configuration of the ForwardRelay component within their distributed systems.
package forwardrelay

import (
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithAddress configures the network address for the ForwardRelay.
// This function provides an option to set the network address of the ForwardRelay, allowing customization of the destination
// where data will be forwarded. It internally uses the SetAddress method to apply the specified address.
//
// Parameters:
//   - address: The network address to set for the ForwardRelay.
//
// Returns:
//
//	A function conforming to types.Option[types.ForwardRelay[T]], which when called with a ForwardRelay component,
//	sets the specified network address.
func WithTarget[T any](targets ...string) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) {
		fr.SetTargets(targets...)
	}
}

// WithLogger configures the logger for the ForwardRelay.
// This function provides an option to set the logger for the ForwardRelay, enabling customization of the logging behavior
// and output destinations. It internally uses the ConnectLogger method to append the specified logger(s).
//
// Parameters:
//   - logger: The logger(s) to append to the ForwardRelay.
//
// Returns:
//
//	A function conforming to types.Option[types.ForwardRelay[T]], which when called with a ForwardRelay component,
//	appends the specified logger(s).
func WithLogger[T any](logger ...types.Logger) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) {
		fr.ConnectLogger(logger...)
	}
}

// WithInput configures the input conduit for the ForwardRelay.
// This function provides an option to set the input conduit for the ForwardRelay, allowing customization of the source(s)
// from which data will be received. It internally uses the ConnectInput method to append the specified input conduit(s).
//
// Parameters:
//   - input: The input conduit(s) to append to the ForwardRelay.
//
// Returns:
//
//	A function conforming to types.Option[types.ForwardRelay[T]], which when called with a ForwardRelay component,
//	appends the specified input conduit(s).
func WithInput[T any](input ...types.Receiver[T]) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) {
		fr.ConnectInput(input...)
	}
}

// WithTLSConfig configures the TLS settings for the ForwardRelay.
// This function provides an option to set the TLS configuration for the ForwardRelay, enabling secure communication
// with remote endpoints. It internally uses the SetTLSConfig method to apply the specified TLS configuration.
//
// Parameters:
//   - config: The TLS configuration to set for the ForwardRelay.
//
// Returns:
//
//	A function conforming to types.Option[types.ForwardRelay[T]], which when called with a ForwardRelay component,
//	sets the specified TLS configuration.
func WithTLSConfig[T any](config *types.TLSConfig) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) {
		fr.SetTLSConfig(config)
	}
}

// WithPerformanceOptions configures the performance options for the ForwardRelay.
// This function provides an option to set performance-related options for the ForwardRelay, allowing customization
// of various performance aspects such as compression, encryption, etc. It internally uses the SetPerformanceOptions
// method to apply the specified performance options.
//
// Parameters:
//   - perfOptions: The performance options to set for the ForwardRelay.
//
// Returns:
//
//	A function conforming to types.Option[types.ForwardRelay[T]], which when called with a ForwardRelay component,
//	sets the specified performance options.
func WithPerformanceOptions[T any](perfOptions *relay.PerformanceOptions) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) {
		fr.SetPerformanceOptions(perfOptions)
	}
}

// WithSecurityOptions configures the security options for the ForwardRelay.
// This function provides an option to set security-related options for the ForwardRelay, allowing
// customization of various security aspects such as encryption or key management. It internally
// uses the SetSecurityOptions method to apply the specified security options.
//
// Parameters:
//   - secOptions:   A pointer to the SecurityOptions struct specifying the new security settings.
//   - encryptionKey: The encryption key to use (e.g. an AES-GCM key).
//
// Returns:
//
//	A function conforming to types.Option[types.ForwardRelay[T]], which when called with a ForwardRelay
//	component, sets the specified security options and encryption key.
func WithSecurityOptions[T any](secOptions *relay.SecurityOptions, encryptionKey string) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) {
		fr.SetSecurityOptions(secOptions, encryptionKey)
	}
}

// WithComponentMetadata configures a ForwardRelay component with custom metadata, including a name and an identifier.
// This function provides an option to set these metadata properties, which can be used for identification,
// logging, or other purposes where metadata is needed for a ForwardRelay. It uses the SetComponentMetadata method
// internally to apply these settings. If the ForwardRelay's configuration is frozen (indicating that the component
// has started operation and its configuration should no longer be changed), attempting to set metadata
// will result in a panic. This ensures the integrity of component configurations during runtime.
//
// Parameters:
//   - name: The name to set for the ForwardRelay component, used for identification and logging.
//   - id: The unique identifier to set for the ForwardRelay component, used for unique identification across systems.
//
// Returns:
//
//	A function conforming to types.Option[types.ForwardRelay[T]], which when called with a ForwardRelay component,
//	sets the specified name and id in the component's metadata.
func WithComponentMetadata[T any](name string, id string) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) {
		fr.SetComponentMetadata(name, id)
	}
}
