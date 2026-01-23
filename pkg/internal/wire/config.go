package wire

import (
	"bytes"
	"context"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SetComponentMetadata updates the wire metadata.
// Panics if called after Start.
func (w *Wire[T]) SetComponentMetadata(name string, id string) {
	w.requireNotStarted("SetComponentMetadata")

	old := w.componentMetadata
	w.componentMetadata.Name = name
	w.componentMetadata.ID = id

	w.NotifyLoggers(
		types.DebugLevel,
		"SetComponentMetadata",
		"component", w.componentMetadata,
		"event", "SetComponentMetadata",
		"result", "SUCCESS",
		"oldComponentMetadata", old,
		"newComponentMetadata", w.componentMetadata,
	)
}

// SetConcurrencyControl sets buffer size and worker count.
// Panics if called after Start.
func (w *Wire[T]) SetConcurrencyControl(bufferSize int, maxConcurrency int) {
	w.requireNotStarted("SetConcurrencyControl")

	if bufferSize < 0 {
		bufferSize = 0
	}
	if maxConcurrency < 1 {
		maxConcurrency = 1
	}

	oldBufferSize := w.maxBufferSize
	oldMaxConcurrency := w.maxConcurrency

	w.maxBufferSize = bufferSize
	w.maxConcurrency = maxConcurrency

	w.NotifyLoggers(
		types.DebugLevel,
		"SetConcurrencyControl",
		"component", w.componentMetadata,
		"event", "SetConcurrencyControl",
		"result", "SUCCESS",
		"oldMaxBufferSize", oldBufferSize,
		"oldMaxConcurrency", oldMaxConcurrency,
		"newMaxBufferSize", w.maxBufferSize,
		"newMaxConcurrency", w.maxConcurrency,
	)
}

// SetEncoder configures the output encoder.
// Panics if called after Start.
func (w *Wire[T]) SetEncoder(e types.Encoder[T]) {
	w.requireNotStarted("SetEncoder")

	oldEncoder := w.encoder
	w.encoder = e

	w.NotifyLoggers(
		types.DebugLevel,
		"SetEncoder",
		"component", w.componentMetadata,
		"event", "SetEncoder",
		"result", "SUCCESS",
		"oldEncoderAddress", oldEncoder,
		"newEncoderAddress", w.encoder,
	)
}

// SetInsulator configures retry behavior for transform errors.
// Panics if called after Start.
func (w *Wire[T]) SetInsulator(retryFunc func(ctx context.Context, elem T, err error) (T, error), threshold int, interval time.Duration) {
	w.requireNotStarted("SetInsulator")

	oldInsulatorFunc := w.insulatorFunc
	oldRetryInterval := w.retryInterval
	oldRetryThreshold := w.retryThreshold

	w.insulatorFunc = retryFunc
	w.retryInterval = interval
	w.retryThreshold = threshold

	w.NotifyLoggers(
		types.DebugLevel,
		"SetInsulator",
		"component", w.componentMetadata,
		"event", "SetInsulator",
		"result", "SUCCESS",
		"oldInsulatorFunc", oldInsulatorFunc,
		"oldRetryInterval", oldRetryInterval,
		"oldRetryThreshold", oldRetryThreshold,
		"newInsulatorFunc", w.insulatorFunc,
		"newRetryInterval", w.retryInterval,
		"newRetryThreshold", w.retryThreshold,
	)
}

// SetOutputBuffer replaces the output buffer.
// Panics if called after Start.
func (w *Wire[T]) SetOutputBuffer(b bytes.Buffer) {
	w.requireNotStarted("SetOutputBuffer")

	oldOutputBuffer := w.OutputBuffer
	w.OutputBuffer = &b

	w.NotifyLoggers(
		types.DebugLevel,
		"SetOutputBuffer",
		"component", w.componentMetadata,
		"event", "SetOutputBuffer",
		"result", "SUCCESS",
		"oldOutputBufferAddress", oldOutputBuffer,
		"newOutputBufferAddress", w.OutputBuffer,
	)
}

// SetErrorChannel replaces the error channel.
// Panics if called after Start.
func (w *Wire[T]) SetErrorChannel(errorChan chan types.ElementError[T]) {
	w.requireNotStarted("SetErrorChannel")

	oldErrorChan := w.errorChan
	w.errorChan = errorChan

	w.NotifyLoggers(
		types.DebugLevel,
		"SetErrorChannel",
		"component", w.componentMetadata,
		"event", "SetErrorChannel",
		"result", "SUCCESS",
		"oldErrorChannelAddress", oldErrorChan,
		"newErrorChannelAddress", w.errorChan,
	)
}

// SetInputChannel replaces the input channel.
// Panics if called after Start.
func (w *Wire[T]) SetInputChannel(inChan chan T) {
	w.requireNotStarted("SetInputChannel")

	oldInputChan := w.inChan
	w.inChan = inChan

	w.NotifyLoggers(
		types.DebugLevel,
		"SetInputChannel",
		"component", w.componentMetadata,
		"event", "SetInputChannel",
		"result", "SUCCESS",
		"oldInputChannelAddress", oldInputChan,
		"newInputChannelAddress", w.inChan,
	)
}

// SetOutputChannel replaces the output channel.
// Panics if called after Start.
func (w *Wire[T]) SetOutputChannel(outChan chan T) {
	w.requireNotStarted("SetOutputChannel")

	oldOutputChan := w.OutputChan
	w.OutputChan = outChan

	w.NotifyLoggers(
		types.DebugLevel,
		"SetOutputChannel",
		"component", w.componentMetadata,
		"event", "SetOutputChannel",
		"result", "SUCCESS",
		"oldOutputChannelAddress", oldOutputChan,
		"newOutputChannelAddress", w.OutputChan,
	)
}

// SetTransformerFactory configures a per-worker transformer factory.
// Panics if called after Start.
func (w *Wire[T]) SetTransformerFactory(factory func() types.Transformer[T]) {
	w.requireNotStarted("SetTransformerFactory")

	if factory == nil {
		w.transformerFactory = nil
		return
	}

	if len(w.transformations) != 0 {
		panic("wire: cannot use TransformerFactory with ConnectTransformer; wrap transforms inside the factory closure instead")
	}

	w.transformerFactory = factory
}
