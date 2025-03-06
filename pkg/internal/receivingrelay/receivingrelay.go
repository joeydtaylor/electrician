// Package receivingrelay provides the implementation for a ReceivingRelay in a distributed data processing system.
// It handles incoming data from network sources or forward relays, processing and routing this data
// to configured outputs. The package includes support for secure communication via TLS, extensive logging
// capabilities for operational insight, and robust error handling mechanisms.
//
// The ReceivingRelay acts as a foundational component in the architecture, enabling reliable and secure
// data exchange within the system. It is designed to be flexible and configurable, allowing it to be
// tailored to specific needs through its initialization options.
//
// This package is essential for developers looking to integrate robust data reception capabilities into their
// distributed systems, providing the tools necessary to receive, process, and forward data efficiently and securely.

package receivingrelay

import (
	"context"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// ReceivingRelay represents a component in a distributed system that receives data from forward relays
// or similar data sources. It encapsulates network listening, TLS configuration, and data distribution
// through a channel-based mechanism to downstream consumers. This structure manages network connections,
// data flow control, and error logging, making it a critical part of the data processing infrastructure.
//
// Fields:
//   - UnimplementedRelayServiceServer: Ensures all RPC methods are implemented, even if not by this struct.
//   - ctx: The context that governs cancellation and lifetime of the relay.
//   - cancel: A function to call to cancel the context and clean up resources.
//   - Outputs: A slice of Submitters to which processed data is sent.
//   - componentMetadata: Metadata describing this relay, such as its unique identifier and type.
//   - Address: The network address where the relay listens for incoming data.
//   - DataCh: A channel through which received data is passed to the processing functions.
//   - Loggers: A slice of Logger instances used for logging operational events.
//   - loggersLock: A mutex that protects access to the Loggers slice.
//   - TlsConfig: Configuration settings for TLS, securing data transmissions.
//   - isRunning: An atomic indicator of whether the relay is actively running.
type ReceivingRelay[T any] struct {
	relay.UnimplementedRelayServiceServer // Embedding the unimplemented server
	ctx                                   context.Context
	cancel                                context.CancelFunc
	Outputs                               []types.Submitter[T]
	componentMetadata                     types.ComponentMetadata // Metadata of the wire.
	Address                               string                  // Address on which the relay listens
	DataCh                                chan T                  // Channel to stream received data
	Loggers                               []types.Logger          // List of loggers attached to the wire.
	loggersLock                           *sync.Mutex
	TlsConfig                             *types.TLSConfig
	isRunning                             int32 // Atomic, use 0 or 1 to represent
	configFrozen                          int32 // Indicates whether the wire's configuration has been frozen, using atomic for thread safety.
}

// NewReceivingRelay initializes and returns a new instance of ReceivingRelay with configuration
// options applied. This constructor method provides a way to configure the relay with custom settings
// such as logging, TLS configuration, and data handling mechanisms before it starts operation.
//
// This function is crucial for setting up a ReceivingRelay that is tailored to specific operational
// requirements, ensuring that all configurations are applied before the relay begins processing data.
//
// Parameters:
//   - ctx: The parent context from which the relay's context is derived.
//   - options: A variadic slice of configuration options that customize the relay's behavior and setup.
//
// Returns:
//   - A configured instance of ReceivingRelay[T] ready for operation.
func NewReceivingRelay[T any](ctx context.Context, options ...types.Option[types.ReceivingRelay[T]]) types.ReceivingRelay[T] {
	ctx, cancel := context.WithCancel(ctx)
	rr := &ReceivingRelay[T]{
		ctx:    ctx,
		cancel: cancel,
		DataCh: make(chan T),
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "RECEIVING_RELAY",
		},
		Loggers:     make([]types.Logger, 0),
		loggersLock: new(sync.Mutex),
	}

	// Apply all provided options to configure the ReceivingRelay.
	for _, option := range options {
		option(rr) // Apply each option directly to the ReceivingRelay instance.
	}

	// Automatically start handling data if an outputConduit is configured.
	if rr.Outputs != nil {
		go func() {
			for data := range rr.DataCh {
				for _, output := range rr.Outputs {
					output.Submit(ctx, data)
				}
			}
		}()
	}

	return rr
}
