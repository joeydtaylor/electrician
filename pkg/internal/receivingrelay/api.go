// Package receivingrelay encapsulates the functionalities required to operate a receiving relay within a distributed system.
// This package is designed to facilitate the secure and efficient reception of data transmitted over the network,
// processing it as per configured rules, and then dispatching it to the next stages of the data pipeline.

// The ReceivingRelay struct is the core component provided by this package. It offers robust networking capabilities,
// including optional TLS-based encryption, to ensure secure data transmissions. The relay supports multiple output
// channels, allowing it to forward processed data to various consumers simultaneously. This design is critical for
// building scalable, reliable, and secure systems that require data ingestion from multiple sources.

// In addition to the relay functionality, this package also provides essential utilities for data handling,
// such as payload unwrapping and decompression, leveraging various compression algorithms to optimize network usage.
// The integration of gRPC provides a powerful, efficient method of data communication, ensuring compatibility and
// performance in microservices architectures.

// This package is crucial for developers implementing back-end systems that need to receive, process, and distribute
// large volumes of data with reliability and security. It offers customizable options to tailor the relay’s behavior
// to specific needs, such as dynamic resource allocation, error handling, and logging for better manageability and observability.

// Key features include:
// - Support for multiple compression algorithms to optimize data transfer.
// - Secure communication channels with configurable TLS settings.
// - Extensive logging capabilities to provide insights into the relay's operations and health.
// - Flexible data channel management to support high throughput and efficient data processing.

// The api.go file specifically contains the implementations of the ReceivingRelay methods, including network operations,
// data handling, and configuration management, providing a comprehensive toolkit for building advanced data ingestion and processing solutions.

package receivingrelay

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ConnectLogger attaches one or more loggers to the ReceivingRelay. This method is useful for enabling
// detailed logging of the relay's operations, which aids in debugging and monitoring.
func (rr *ReceivingRelay[T]) ConnectLogger(loggers ...types.Logger) {
	if atomic.LoadInt32(&rr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=ConnectLogger", rr.componentMetadata))
	}
	rr.Loggers = append(rr.Loggers, loggers...)
	for _, l := range loggers {
		rr.NotifyLoggers(types.DebugLevel, "component: %s, address: %s,level: DEBUG, result: SUCCESS, event: ConnectLogger, target: %v => ConnectLogger called", rr.componentMetadata, rr.Address, l)
	}
}

// ConnectOutput connects one or more output Submitters to the ReceivingRelay. This setup allows
// the ReceivingRelay to pass processed data to subsequent stages in a data pipeline.
func (rr *ReceivingRelay[T]) ConnectOutput(outputs ...types.Submitter[T]) {
	for _, out := range outputs {
		rr.NotifyLoggers(types.DebugLevel, "component: %s, address: %s,level: DEBUG, result: SUCCESS, event: ConnectOutput, target: %v => ConnectOutput called", rr.componentMetadata, rr.Address, out.GetComponentMetadata())
	}
	rr.Outputs = append(rr.Outputs, outputs...)
}

// GetAddress retrieves the network address at which the ReceivingRelay is configured to listen.
// This method aids in identifying the network configuration of the relay.
func (rr *ReceivingRelay[T]) GetAddress() string {
	rr.NotifyLoggers(types.DebugLevel, "component: %s, address: %s,level: DEBUG, result: SUCCESS, event: GetAddress, return: %v => GetAddress called", rr.componentMetadata, rr.Address, rr.Address)
	return rr.Address
}

// GetComponentMetadata retrieves the metadata of the ReceivingRelay, which includes identifiers
// and other descriptive information that may be useful for logging or managing components.
func (rr *ReceivingRelay[T]) GetComponentMetadata() types.ComponentMetadata {
	rr.NotifyLoggers(types.DebugLevel, "component: %s, address: %s,level: DEBUG, result: SUCCESS, event: GetComponentMetadata, return: %v => GetComponentMetadata called", rr.componentMetadata, rr.Address, rr.componentMetadata)
	return rr.componentMetadata
}

// GetDataChannel returns the channel used by the ReceivingRelay to receive data internally.
// This can be useful for debugging or integrating custom processing logic.
func (rr *ReceivingRelay[T]) GetDataChannel() chan T {
	rr.NotifyLoggers(types.DebugLevel, "component: %s, address: %s,level: DEBUG, result: SUCCESS, event: GetDataChannel, return: %v => GetDataChannel called", rr.componentMetadata, rr.Address, rr.DataCh)
	return rr.DataCh
}

// GetOutputs retrieves a list of all output Submitters connected to the ReceivingRelay.
// This allows for inspection and modification of the relay's output behavior.
func (rr *ReceivingRelay[T]) GetOutputs() []types.Submitter[T] {
	rr.NotifyLoggers(types.DebugLevel, "component: %s, address: %s,level: DEBUG, result: SUCCESS, event: GetOutputs, return: %v => GetOutputs called", rr.componentMetadata, rr.Address, rr.Outputs)
	return rr.Outputs
}

// GetOutputChannel provides the output channel for the ReceivingRelay, allowing other components
// to receive processed data from this relay.
func (rr *ReceivingRelay[T]) GetOutputChannel() chan T {
	rr.NotifyLoggers(types.DebugLevel, "%s, address: %s,level: DEBUG, result: SUCCESS, event: GetOutputChannel, address: %s, return: %v => GetOutputChannel called", rr.componentMetadata, rr.Address, rr.Address, rr.DataCh)
	return rr.DataCh
}

// GetTLSConfig retrieves the TLS configuration of the ReceivingRelay, which includes settings
// necessary for secure communications.
func (rr *ReceivingRelay[T]) GetTLSConfig() *types.TLSConfig {
	rr.NotifyLoggers(types.DebugLevel, "component: %s, address: %s,level: DEBUG, result: SUCCESS, event: GetTLSConfig, target: %v => GetTLSConfig called", rr.componentMetadata, rr.Address, rr.TlsConfig)
	return rr.TlsConfig
}

// IsRunning checks the operational status of the ReceivingRelay, returning true if it is actively processing data.
func (fr *ReceivingRelay[T]) IsRunning() bool {
	return atomic.LoadInt32(&fr.isRunning) == 1
}

// Listen starts the server process for the ReceivingRelay, enabling it to accept incoming connections
// and process data according to the configured logic and network settings.
func (rr *ReceivingRelay[T]) Listen(listenForever bool, retryInSeconds int) error {
	var opts []grpc.ServerOption
	if rr.TlsConfig != nil && rr.TlsConfig.UseTLS {
		rr.NotifyLoggers(types.DebugLevel, "%s, address: %s, level: DEBUG, result: SUCCESS, event: Listen => TLS is enabled, loading credentials", rr.componentMetadata, rr.Address)
		creds, err := rr.loadTLSCredentials(rr.TlsConfig)
		if err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "%s, address: %s, level: ERROR, result: FAILURE, event: Listen, => Failed to load TLS credentials: %v", rr.componentMetadata, rr.Address, err)
			return fmt.Errorf("failed to load TLS credentials: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
		rr.NotifyLoggers(types.DebugLevel, "%s, address: %s, level: DEBUG, result: SUCCESS, event: Listen => TLS credentials loaded", rr.componentMetadata, rr.Address)
	} else {
		rr.NotifyLoggers(types.DebugLevel, "%s, address: %s, level: DEBUG, result: SUCCESS, event: Listen => TLS is not enabled", rr.componentMetadata, rr.Address)
	}

	rr.NotifyLoggers(types.InfoLevel, "%s, address: %s, level: INFO, result: SUCCESS, event: Listen => Attempting to listen at: %s", rr.componentMetadata, rr.Address, rr.Address)
	lis, err := net.Listen("tcp", rr.Address)
	if err != nil {
		rr.NotifyLoggers(types.ErrorLevel, "%s, address: %s, level: ERROR, result: FAILURE, event: Listen, => Failed to listen at %v: %v", rr.componentMetadata, rr.Address, rr.Address, err)
		if !listenForever {
			return err
		}
		rr.NotifyLoggers(types.WarnLevel, "%s, address: %s, level: WARN, result: SUCCESS, event: Listen => Retrying listen after %d seconds", rr.componentMetadata, rr.Address, retryInSeconds)
		time.Sleep(time.Duration(retryInSeconds) * time.Second)
		return rr.Listen(listenForever, retryInSeconds)
	}

	rr.NotifyLoggers(types.InfoLevel, "%s, address: %s, level: INFO, result: SUCCESS, event: Listen => TCP listener created, starting gRPC server", rr.componentMetadata, rr.Address)
	s := grpc.NewServer(opts...)
	relay.RegisterRelayServiceServer(s, rr)

	rr.NotifyLoggers(types.InfoLevel, "%s, address: %s, level: INFO, result: SUCCESS, event: Listen, => gRPC server registered, beginning to serve at address: %s", rr.componentMetadata, rr.Address, rr.Address)
	if !listenForever {
		if err := s.Serve(lis); err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "%s, address: %s, level: ERROR, result: FAILURE, event: Listen, error: %v => gRPC server stopped:", rr.componentMetadata, rr.Address, err)
		}
		return nil
	}

	go func() {
		if err := s.Serve(lis); err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "%s, address: %s, level: ERROR, result: FAILURE, event: Listen, error: %v => gRPC server failed:", rr.componentMetadata, rr.Address, err)
		}
	}()

	rr.NotifyLoggers(types.InfoLevel, "%s, address: %s, level: INFO, result: SUCCESS, event: Listen => gRPC server started", rr.componentMetadata, rr.Address)
	return nil
}

// NotifyLoggers sends a formatted log message to all attached loggers, facilitating unified logging
// across various components of the ReceivingRelay.
func (rr *ReceivingRelay[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if rr.Loggers != nil {
		msg := fmt.Sprintf(format, args...)
		for _, logger := range rr.Loggers {
			if logger == nil {
				continue // Skip if the logger is nil.
			}
			rr.loggersLock.Lock()
			defer rr.loggersLock.Unlock()
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

// Receive processes incoming data payloads within a context, providing acknowledgments upon successful
// processing and facilitating error handling.
func (rr *ReceivingRelay[T]) Receive(ctx context.Context, payload *relay.WrappedPayload) (*relay.StreamAcknowledgment, error) {
	// Extract or generate a trace ID
	md, ok := metadata.FromIncomingContext(ctx)
	var traceID string
	if !ok || len(md["trace-id"]) == 0 {
		traceID = utils.GenerateUniqueHash() // Generate a new trace ID if not received
	} else {
		traceID = md["trace-id"][0] // Use existing trace ID from metadata
	}

	rr.NotifyLoggers(types.InfoLevel, "%s, address: %s, level: INFO, result: SUCCESS, event: Receive, trace_id: %v => Received stream.", rr.componentMetadata, rr.Address, traceID)
	ack := &relay.StreamAcknowledgment{Success: true, Message: "Received stream"}
	go func() {
		var data T
		if err := UnwrapPayload(payload, &data); err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "%s, address: %s, level: ERROR, result: FAILURE, event: Receive, error: %v, trace_id: %v => Error unwrapping payload", rr.componentMetadata, rr.Address, err, traceID)
			return
		}
		rr.NotifyLoggers(types.InfoLevel, "%s, address: %s, level: INFO, result: SUCCESS, event: Receive, data: %v, trace_id: %v => Data unwrapped and sent to channel", rr.componentMetadata, rr.Address, data, traceID)
		rr.DataCh <- data
	}()
	return ack, nil
}

// SetAddress configures the network address for the ReceivingRelay. This is crucial for network-based
// operations and must be set prior to starting the relay.
func (rr *ReceivingRelay[T]) SetAddress(address string) {
	if atomic.LoadInt32(&rr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetAddress", rr.componentMetadata))
	}
	rr.NotifyLoggers(types.DebugLevel, "component: %s, address: %s,level: DEBUG, result: SUCCESS, event: SetAddress, new: %v => SetAddress called", rr.componentMetadata, rr.Address, address)
	rr.Address = address
}

// SetDataChannel configures the buffer size for the internal data channel of the ReceivingRelay, impacting
// how data is buffered during processing.
func (rr *ReceivingRelay[T]) SetDataChannel(bufferSize uint32) {
	if atomic.LoadInt32(&rr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetDataChannel", rr.componentMetadata))
	}
	rr.DataCh = make(chan T, bufferSize)
	rr.NotifyLoggers(types.DebugLevel, "component: %s, address: %s, level: DEBUG, result: SUCCESS, event: SetDataChannel, new: %v => SetDataChannel called", rr.componentMetadata, rr.Address, rr.DataCh)
}

// SetComponentMetadata allows for setting or updating the metadata of the ReceivingRelay, such as its name
// and unique identifier, which is crucial for component management and identification.
func (rr *ReceivingRelay[T]) SetComponentMetadata(name string, id string) {
	if atomic.LoadInt32(&rr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetComponentMetadata", rr.componentMetadata))
	}
	old := rr.componentMetadata
	rr.componentMetadata.Name = name
	rr.componentMetadata.ID = id
	rr.NotifyLoggers(types.DebugLevel, "component: %s, address: %s,level: DEBUG, result: SUCCESS, event: SetComponentMetadata, new: => SetComponentMetadata called", old, rr.componentMetadata)
}

// SetTLSConfig configures the TLS settings for the ReceivingRelay, securing its communication channels
// according to the specified TLS parasensors.
func (rr *ReceivingRelay[T]) SetTLSConfig(config *types.TLSConfig) {
	if atomic.LoadInt32(&rr.configFrozen) == 1 {
		panic(fmt.Sprintf("attempted to modify frozen configuration of started component: %s, action=SetTLSConfig", rr.componentMetadata))
	}
	rr.TlsConfig = config
	rr.NotifyLoggers(types.DebugLevel, "component: %s, address: %s,level: DEBUG, result: SUCCESS, event: SetTLSConfig, new: %v => SetTLSConfig called", rr.componentMetadata, rr.Address, rr.TlsConfig)
}

// Start initiates the operations of the ReceivingRelay, making it begin its processing and networking functions.
func (rr *ReceivingRelay[T]) Start(ctx context.Context) error {
	atomic.StoreInt32(&rr.configFrozen, 1)
	rr.NotifyLoggers(types.InfoLevel, "component: %s, address: %s, level: INFO, result: SUCCESS, event: Start => Starting Receiving Relay", rr.componentMetadata, rr.Address)
	for _, output := range rr.Outputs {
		if !output.IsStarted() {
			output.Start(ctx)
		}
	}
	go rr.Listen(true, 0)
	atomic.StoreInt32(&rr.isRunning, 1)

	return nil
}

// StreamReceive handles a continuous stream of data from a client, processing each piece of data as it arrives
// and maintaining a session through the provided server stream interface.
func (rr *ReceivingRelay[T]) StreamReceive(stream relay.RelayService_StreamReceiveServer) error {
	for {
		rr.NotifyLoggers(types.InfoLevel, "%s, address: %s, level: INFO, result: SUCCESS, event: StreamReceive => Waiting to receive stream data", rr.componentMetadata, rr.Address)
		payload, err := stream.Recv()
		if err == io.EOF {
			rr.NotifyLoggers(types.InfoLevel, "%s, address: %s, level: INFO, result: SUCCESS, event: StreamReceive => End of data stream", rr.componentMetadata, rr.Address)
			return nil
		}
		if err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "%s, address: %s, level: ERROR, result: FAILURE, event: StreamReceive, error: %v => Error receiving stream data", rr.componentMetadata, rr.Address, err)
			return err
		}

		// Extract or generate a trace ID
		md, _ := metadata.FromIncomingContext(stream.Context()) // Assuming the stream's context carries metadata
		var traceID string
		if len(md["trace-id"]) == 0 {
			traceID = utils.GenerateUniqueHash() // Generate a new trace ID if not received
		} else {
			traceID = md["trace-id"][0] // Use existing trace ID from metadata
		}

		var data T
		if err = UnwrapPayload(payload, &data); err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "%s, address: %s, level: ERROR, result: FAILURE, event: StreamReceive, trace_id: %v => Failed to unwrap payload: %s", rr.componentMetadata, rr.Address, traceID, err)
			return fmt.Errorf("failed to unwrap payload: %v", err)
		}
		rr.NotifyLoggers(types.InfoLevel, "%s, address: %s, level: INFO, event: StreamReceive, trace_id: %v => Streaming data received: %v", rr.componentMetadata, rr.Address, traceID, data)
		rr.DataCh <- data

		// Create an acknowledgment for each received message with the trace ID
		ack := &relay.StreamAcknowledgment{Success: true, Message: "Data received successfully"}
		if err := stream.Send(ack); err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "%s, address: %s, level: ERROR, result: FAILURE, event: StreamReceive, error: %v, trace_id: %v => Failed to send acknowledgment:", rr.componentMetadata, rr.Address, err, traceID)
			return fmt.Errorf("failed to send acknowledgment: %v", err)
		}
	}
}

// Terminate stops all operations of the ReceivingRelay, ensuring all resources are properly released and
// any persistent connections are closed gracefully.
func (rr *ReceivingRelay[T]) Stop() {
	rr.NotifyLoggers(types.InfoLevel, "component: %s, address: %s, level: INFO, result: SUCCESS, event: Terminate => Terminating Receiving Relay", rr.componentMetadata, rr.Address)

	// Signal all processes to stop
	rr.cancel()

	// Safely close the DataCh
	close(rr.DataCh)

	// Ensure all outputs are properly terminated
	for _, output := range rr.Outputs {
		output.Stop()
	}

	// Indicate the relay has fully stopped
	atomic.StoreInt32(&rr.isRunning, 0)
}

const (
	COMPRESS_NONE    relay.CompressionAlgorithm = 0
	COMPRESS_DEFLATE relay.CompressionAlgorithm = 1
	COMPRESS_SNAPPY  relay.CompressionAlgorithm = 2
	COMPRESS_ZSTD    relay.CompressionAlgorithm = 3
	COMPRESS_BROTLI  relay.CompressionAlgorithm = 4
	COMPRESS_LZ4     relay.CompressionAlgorithm = 5
)

// UnwrapPayload decodes the data contained in a WrappedPayload into the specified generic type T, handling
// potential compression and format differences based on the relay's configuration.
func UnwrapPayload[T any](wrappedPayload *relay.WrappedPayload, data *T) error {
	if wrappedPayload.Metadata == nil || wrappedPayload.Metadata.Performance == nil {
		return errors.New("metadata or performance options are nil")
	}

	var buf *bytes.Buffer

	// Check if compression was used and decompress accordingly
	if wrappedPayload.Metadata.Performance.UseCompression {
		var err error
		buf, err = decompressData(wrappedPayload.Payload, wrappedPayload.Metadata.Performance.CompressionAlgorithm)
		if err != nil {
			return err
		}
	} else {
		buf = bytes.NewBuffer(wrappedPayload.Payload)
	}

	// Decode the payload using gob
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(data); err != nil {
		return err
	}

	return nil
}
