// Package forwardrelay implements a sophisticated component designed for forwarding data with high reliability and configurability,
// leveraging modern protocols and techniques in a distributed system architecture. This package's api.go file serves as a central
// hub for defining the API operations of the ForwardRelay component, encapsulating functionalities necessary for initializing,
// configuring, and managing the lifecycle and data flow of the relay.

// The api.go file in the ForwardRelay package includes the implementation of various crucial methods:
// - Connection Methods: Methods such as ConnectInput and ConnectLogger, which are crucial for setting up the data inputs and logging mechanisms,
//   ensuring that the relay can receive data streams and appropriately log operations for debugging and monitoring.
// - Getter Methods: These methods, including GetAddress, GetComponentMetadata, and others, provide read access to the relay's configuration
//   and state, facilitating easy integration with other components and systems by exposing essential information.
// - Setter Methods: Including SetAddress, SetComponentMetadata, and SetTLSConfig, these methods allow for the dynamic configuration of the relay
//   during runtime, adapting to changes in the operational environment or requirements.
// - Operational Methods: Start, Submit, and Terminate control the operational state of the relay, handling the starting, data processing,
//   and graceful shutdown of the component.
// - Utility Methods: NotifyLoggers and WrapPayload, which support the core functionalities by providing logging capabilities and data packaging
//   for network transmission, respectively.

// Together, these functionalities make the ForwardRelay a robust tool for data transmission in distributed systems, supporting detailed configuration,
// dynamic management, and high reliability.
package forwardrelay

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConnectInput appends one or more data receivers to the ForwardRelay.
// This function is crucial for setting up the data ingestion points for the relay,
// enabling it to receive and forward data from configured sources.
// Each receiver is registered, and a debug log is generated to confirm the operation.
//
// Parameters:
//   - r: A variadic slice of Receiver[T] interfaces, representing the data sources to be connected.
//
// This method logs each connection attempt, including the metadata of the connected receivers,
// providing traceability and ease of monitoring the connectivity status.
func (fr *ForwardRelay[T]) ConnectInput(r ...types.Receiver[T]) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=ConnectInput", fr.componentMetadata))
	}
	for _, i := range r {
		fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: ConnectInput, target: %v => ConnectInput called", fr.componentMetadata, i.GetComponentMetadata())
	}
	fr.Input = append(fr.Input, r...)
}

// ConnectLogger appends one or more logger instances to the ForwardRelay.
// Loggers are essential for monitoring the operations of the relay, capturing events, and aiding in debugging.
// This method enhances operational transparency and aids in the diagnostic process.
//
// Parameters:
//   - logger: A variadic slice of Logger interfaces, which are to be connected to the relay for logging purposes.
//
// Each logger is registered, and a debug message is logged to confirm the successful attachment of each logger,
// ensuring that all operational states and changes are recorded.
func (fr *ForwardRelay[T]) ConnectLogger(logger ...types.Logger) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=ConnectLogger", fr.componentMetadata))
	}
	fr.Loggers = append(fr.Loggers, logger...)
	for _, l := range logger {
		fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: ConnectLogger, target: %v => ConnectLogger called", fr.componentMetadata, l)
	}
}

// GetAddress retrieves the network address at which the ForwardRelay is configured to listen.
// This is essential for network configurations and connectivity checks.
//
// Returns:
//   - string: The network address of the ForwardRelay.
//
// The method logs the retrieval action at the debug level, providing an audit trail for configuration access.
func (fr *ForwardRelay[T]) GetTargets() []string {
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: GetAddress, return: %v => GetAddress called", fr.componentMetadata, fr.Targets)
	return fr.Targets
}

// GetComponentMetadata retrieves the metadata associated with the ForwardRelay.
// Metadata includes the component's identification and type, which are crucial for system integration and monitoring.
//
// Returns:
//   - types.ComponentMetadata: The metadata object of the ForwardRelay.
//
// This method logs the action of accessing the component metadata at the debug level to ensure traceability.
func (fr *ForwardRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: GetComponentMetadata, return: %v => GetComponentMetadata called", fr.componentMetadata, fr.componentMetadata)
	return fr.componentMetadata
}

// GetInput retrieves the list of data receivers currently connected to the ForwardRelay.
// This information is key to understanding the data inputs and integration points of the relay.
//
// Returns:
//   - []types.Receiver[T]: A slice of Receiver interfaces currently receiving data for the relay.
//
// This method logs the retrieval of connected receivers at the debug level, enhancing operational transparency.
func (fr *ForwardRelay[T]) GetInput() []types.Receiver[T] {
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: GetInput, return: %v => GetInput called", fr.componentMetadata, fr.Input)
	return fr.Input
}

// GetPerformanceOptions retrieves the performance optimization settings of the ForwardRelay.
// These settings dictate how the relay processes and transmits data, including compression and other performance-related options.
//
// Returns:
//   - *relay.PerformanceOptions: A pointer to the PerformanceOptions struct, providing access to current settings like compression and batch sizes.
//
// The function logs the access to performance settings at a debug level to ensure that any access or retrieval of these settings is traceable and transparent.
func (fr *ForwardRelay[T]) GetPerformanceOptions() *relay.PerformanceOptions {
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: GetPerformanceOptions, return: %v => GetPerformanceOptions called", fr.componentMetadata, fr.PerformanceOptions)
	return fr.PerformanceOptions
}

// GetTLSConfig retrieves the TLS configuration used by the ForwardRelay.
// This configuration is crucial for establishing secure connections and handling data confidentiality and integrity during transmission.
//
// Returns:
//   - *types.TLSConfig: A pointer to the TLSConfig struct that contains settings like certificates and keys for secure communication.
//
// This method logs the retrieval of TLS settings at the debug level, providing a record of security configuration access, which is essential for security audits.
func (fr *ForwardRelay[T]) GetTLSConfig() *types.TLSConfig {
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: GetTLSConfig, return: %v => GetTLSConfig called", fr.componentMetadata, fr.TlsConfig)
	return fr.TlsConfig
}

// IsRunning checks the operational status of the ForwardRelay.
// This function is essential for monitoring and management, ensuring that the system's operational status is always known.
//
// Returns:
//   - bool: True if the relay is currently active and running, false otherwise.
//
// This method provides a thread-safe check of the relay's state, using atomic operations to ensure that the returned status is current and accurate.
func (fr *ForwardRelay[T]) IsRunning() bool {
	return atomic.LoadInt32(&fr.isRunning) == 1
}

// SetAddress updates the network address at which the ForwardRelay listens for incoming connections.
// Changing the address can be crucial for reconfiguring the relay in response to network changes or operational requirements.
//
// Parameters:
//   - address: The new network address to be set for the relay.
//
// This method logs the change of address at a debug level, ensuring that all modifications to critical configuration details are recorded and traceable.
func (fr *ForwardRelay[T]) SetTargets(address ...string) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetAddress", fr.componentMetadata))
	}
	fr.Targets = append(fr.Targets, address...)
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: SetAddress, new: %v => SetAddress called", fr.componentMetadata, fr.Targets)
}

// SetComponentMetadata updates the metadata of the ForwardRelay, such as its name and unique identifier.
// This metadata is crucial for identifying and managing the relay in a distributed system or within larger infrastructure.
//
// Parameters:
//   - name: The new name to be set for the relay.
//   - id: The new unique identifier to be set for the relay.
//
// This method logs the metadata change at a debug level, capturing the previous state and new values,
// thus providing traceability and aiding in debugging and management of relay components.
func (fr *ForwardRelay[T]) SetComponentMetadata(name string, id string) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetComponentMetadata", fr.componentMetadata))
	}
	old := fr.componentMetadata
	fr.componentMetadata.Name = name
	fr.componentMetadata.ID = id
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: SetComponentMetadata, new: %v => SetComponentMetadata called", old, fr.componentMetadata)
}

// SetPerformanceOptions updates the performance settings for the ForwardRelay.
// These settings affect how the relay handles data processing tasks such as compression and batching,
// which can optimize throughput and efficiency.
//
// Parameters:
//   - perfOptions: A pointer to the PerformanceOptions struct specifying the new performance settings.
//
// This method logs the update of performance settings at a debug level, ensuring that changes to critical performance-related configurations are recorded for auditing and performance tuning.
func (fr *ForwardRelay[T]) SetPerformanceOptions(perfOptions *relay.PerformanceOptions) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetPerformanceOptions", fr.componentMetadata))
	}
	fr.PerformanceOptions = perfOptions
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: SetPerformanceOptions, new: %v => SetPerformanceOptions called", fr.componentMetadata, fr.PerformanceOptions)
}

// SetSecurityOptions updates the security settings for the ForwardRelay.
// This includes whether encryption is enabled (e.g., AES-GCM), and also stores
// an out-of-band encryption key for use in securing data.
//
// Parameters:
//   - secOptions: A pointer to the SecurityOptions struct specifying the new security settings.
//   - encryptionKey: The encryption key to use (e.g. an AES-GCM key).
//
// This method logs the update of security settings at a debug level, ensuring changes
// to critical security-related configurations are recorded for auditing and compliance.
func (fr *ForwardRelay[T]) SetSecurityOptions(secOptions *relay.SecurityOptions, encryptionKey string) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetSecurityOptions", fr.componentMetadata))
	}
	fr.SecurityOptions = secOptions
	fr.EncryptionKey = encryptionKey

	fr.NotifyLoggers(types.DebugLevel,
		"component: %v, level: DEBUG, result: SUCCESS, event: SetSecurityOptions, new: %v => SetSecurityOptions called",
		fr.componentMetadata, fr.SecurityOptions)
}

// SetTLSConfig updates the TLS (Transport Layer Security) configuration of the ForwardRelay.
// This configuration is essential for ensuring secure communication channels in environments where data integrity and confidentiality are paramount.
//
// Parameters:
//   - config: A pointer to the TLSConfig struct that contains the new TLS settings such as certificates and keys.
//
// This method logs the TLS configuration update at a debug level, providing a traceable record of how security settings are modified, which is crucial for maintaining the security posture of the system.
func (fr *ForwardRelay[T]) SetTLSConfig(config *types.TLSConfig) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetTLSConfig", fr.componentMetadata))
	}
	fr.TlsConfig = config
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: SetTLSConfig, new: %v => SetTLSConfig called", fr.componentMetadata, fr.TlsConfig)
}

// Start initiates the operations of the ForwardRelay, making it begin its processing and networking functions.
// It verifies the initialization of essential components such as Input and Loggers before proceeding.
// This method also starts the receivers that are connected to this relay, ensuring they are ready to handle incoming data.
//
// Start logs at an informational level when the relay begins processing, and updates its running state atomically to indicate it's operational.
func (fr *ForwardRelay[T]) Start(ctx context.Context) error {
	if fr.Input == nil {
		return fmt.Errorf("no inputs configured")
	}

	atomic.StoreInt32(&fr.configFrozen, 1)

	for _, input := range fr.Input {
		if !input.IsStarted() {
			if err := input.Start(ctx); err != nil {
				return fmt.Errorf("failed to start input %v: %w", input.GetComponentMetadata(), err)
			}
		}
		go fr.readFromInput(input)
	}

	atomic.StoreInt32(&fr.isRunning, 1)
	fr.NotifyLoggers(types.InfoLevel, "component: %v, level: INFO, event: Start => forward relay running", fr.componentMetadata)
	return nil
}

// Submit handles the transmission of items via the ForwardRelay to an external receiver.
// This method involves several critical steps:
// 1. Extracting or generating a trace ID from the context for tracking and logging.
// 2. Wrapping the item in a structured payload (gob-encoded, optional compression, optional encryption).
// 3. Setting up and using a gRPC client to send the wrapped payload, handling TLS if enabled.
// 4. Logging the progress and results of data transmission.
//
// The function returns an error if it encounters issues during wrapping or data transmission.
func (fr *ForwardRelay[T]) Submit(ctx context.Context, item T) error {
	// Wrap the payload (gob -> compress -> encrypt) as before
	wp, err := WrapPayload(item, fr.PerformanceOptions, fr.SecurityOptions, fr.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to wrap payload: %w", err)
	}

	// Assign sequence for correlation
	wp.Seq = atomic.AddUint64(&fr.seq, 1)

	// If stream defaults are set (StreamOpen.defaults), omit per-message metadata to reduce overhead
	wp.Metadata = nil

	env := &relay.RelayEnvelope{
		Msg: &relay.RelayEnvelope_Payload{Payload: wp},
	}

	// Send to each target stream (enqueue, don’t dial per item)
	for _, address := range fr.Targets {
		sess, err := fr.getOrCreateStreamSession(ctx, address)
		if err != nil {
			fr.NotifyLoggers(types.ErrorLevel,
				"component: %v, level: ERROR, event: Submit => stream setup failed addr=%s err=%v",
				fr.componentMetadata, address, err,
			)
			continue
		}

		select {
		case sess.sendCh <- env:
			// enqueued
		case <-ctx.Done():
			return ctx.Err()
		case <-fr.ctx.Done():
			return fr.ctx.Err()
		}
	}

	return nil
}

// NotifyLoggers sends a formatted log message to all attached loggers.
// This method is critical for maintaining detailed logs of the ForwardRelay's operations, aiding in debugging and monitoring.
//
// Parameters:
//   - level: The severity level of the log message.
//   - format: The format string for the log message.
//   - args: Arguments to be formatted into the log message.
//
// This method iterates through all attached loggers and sends them the formatted message if they are initialized.
func (fr *ForwardRelay[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if fr.Loggers != nil {
		msg := fmt.Sprintf(format, args...)
		for _, logger := range fr.Loggers {
			if logger == nil {
				continue // Skip if the logger is nil.
			}
			fr.loggersLock.Lock()
			defer fr.loggersLock.Unlock()
			switch level {
			case types.DebugLevel:
				logger.Debug(msg)
			case types.InfoLevel:
				logger.Info(msg)
			case types.WarnLevel:
				logger.Warn(msg)
			case types.ErrorLevel:
				logger.Error(msg)
			case types.DPanicLevel:
				logger.DPanic(msg)
			case types.PanicLevel:
				logger.Panic(msg)
			case types.FatalLevel:
				logger.Fatal(msg)
			}
		}
	}
}

// Terminate halts all operations of the ForwardRelay and updates its running state.
// This method logs an informational message indicating the relay's termination and ensures all resources are properly released.
//
// It atomically updates the `isRunning` state to ensure that the relay's status is accurately reflected across all threads.
func (fr *ForwardRelay[T]) Stop() {
	fr.NotifyLoggers(types.InfoLevel,
		"component: %s, level: INFO, result: SUCCESS, event: Stop => Stopping Forward Relay",
		fr.componentMetadata,
	)

	// Stop producers first so nothing keeps submitting.
	fr.cancel()

	// Close streams (CloseSend + conn close) and wait.
	fr.closeAllStreams("relay stop")

	atomic.StoreInt32(&fr.isRunning, 0)
}

// WrapPayload serializes the given data of generic type T into a WrappedPayload structure,
// applying compression and optional AES-GCM encryption if enabled. The resulting bytes
// go into the payload of the returned WrappedPayload.
//
// Parameters:
//   - data: The generic data to be serialized.
//   - perfOpts: PerformanceOptions controlling compression.
//   - secOpts: SecurityOptions controlling encryption (suite, enabled).
//   - encryptionKey: The AES-GCM key if encryption is enabled.
//
// Returns:
//   - *relay.WrappedPayload: The newly created payload with compressed/encrypted data
//   - error: If an issue occurs during serialization, compression, or encryption
func WrapPayload[T any](
	data T,
	perfOpts *relay.PerformanceOptions,
	secOpts *relay.SecurityOptions,
	encryptionKey string,
) (*relay.WrappedPayload, error) {

	// Default performance opts if nil
	if perfOpts == nil {
		perfOpts = &relay.PerformanceOptions{UseCompression: false, CompressionAlgorithm: COMPRESS_NONE}
	}

	// 1) Serialize via GOB into buf
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return nil, fmt.Errorf("gob encode failed: %w", err)
	}

	// 2) Compress if enabled
	if perfOpts.UseCompression {
		compressed, err := compressData(buf.Bytes(), perfOpts.CompressionAlgorithm)
		if err != nil {
			return nil, fmt.Errorf("compression failed: %w", err)
		}
		buf.Reset()
		buf.Write(compressed)
	}

	// 3) Encrypt if secOpts says so (AES-GCM)
	if secOpts != nil && secOpts.Enabled && secOpts.Suite == ENCRYPT_AES_GCM {
		encrypted, err := encryptData(buf.Bytes(), secOpts, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		buf.Reset()
		buf.Write(encrypted)
	}

	// 4) Build the final WrappedPayload
	timestamp := timestamppb.Now()
	id := timestamp.AsTime().Format(time.RFC3339Nano)

	metadata := &relay.MessageMetadata{
		Headers: map[string]string{
			"source": "go",
		},
		ContentType: "application/octet-stream",
		Version: &relay.VersionInfo{
			Major: 1,
			Minor: 0,
		},
		Performance: perfOpts,
		Security:    secOpts,
	}

	return &relay.WrappedPayload{
		Id:        id,
		Timestamp: timestamp,
		Payload:   buf.Bytes(),
		Metadata:  metadata,
	}, nil
}

func (fr *ForwardRelay[T]) SetAuthenticationOptions(opts *relay.AuthenticationOptions) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetAuthenticationOptions", fr.componentMetadata))
	}
	fr.authOptions = opts
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: SetAuthenticationOptions, new: %+v", fr.componentMetadata, opts)
}

func (fr *ForwardRelay[T]) SetOAuth2(ts types.OAuth2TokenSource) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetOAuth2", fr.componentMetadata))
	}
	fr.tokenSource = ts
	if ts == nil {
		fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: SetOAuth2, state: disabled", fr.componentMetadata)
	} else {
		fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: SetOAuth2, state: enabled", fr.componentMetadata)
	}
}

func (fr *ForwardRelay[T]) SetStaticHeaders(h map[string]string) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetStaticHeaders", fr.componentMetadata))
	}
	fr.staticHeaders = make(map[string]string, len(h))
	for k, v := range h {
		fr.staticHeaders[k] = v
	}
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: SetStaticHeaders, keys: %d", fr.componentMetadata, len(fr.staticHeaders))
}

func (fr *ForwardRelay[T]) SetDynamicHeaders(fn func(ctx context.Context) map[string]string) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetDynamicHeaders", fr.componentMetadata))
	}
	fr.dynamicHeaders = fn
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: SetDynamicHeaders, installed: %t", fr.componentMetadata, fn != nil)
}

// ForwardRelayWithAuthRequired is an optional builder option to toggle the
// per-RPC bearer requirement at construction time.
func ForwardRelayWithAuthRequired[T any](required bool) types.Option[types.ForwardRelay[T]] {
	return func(c types.ForwardRelay[T]) {
		if fr, ok := c.(*ForwardRelay[T]); ok {
			fr.authRequired = required
		}
	}
}

// SetAuthRequired flips whether a valid bearer token is required before any
// gRPC call when OAuth2 is configured. Must be called before Start().
func (fr *ForwardRelay[T]) SetAuthRequired(required bool) {
	if atomic.LoadInt32(&fr.configFrozen) == 1 {
		panic("attempted to modify frozen configuration of started component: SetAuthRequired")
	}
	fr.authRequired = required
	fr.NotifyLoggers(types.DebugLevel,
		"component: %v, level: DEBUG, result: SUCCESS, event: SetAuthRequired, required: %t",
		fr.componentMetadata, fr.authRequired)
}

// GetAuthRequired reports the current setting.
func (fr *ForwardRelay[T]) GetAuthRequired() bool {
	return fr.authRequired
}
