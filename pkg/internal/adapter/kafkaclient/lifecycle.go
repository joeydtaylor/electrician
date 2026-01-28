package kafkaclient

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// Stop terminates reader and writer activity and emits stop hooks.
func (a *KafkaClient[T]) Stop() {
	for _, sensor := range a.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnKafkaWriterStop(a.componentMetadata)
		sensor.InvokeOnKafkaConsumerStop(a.componentMetadata)
	}
	a.NotifyLoggers(
		types.InfoLevel,
		"Stop",
		"component", a.componentMetadata,
		"event", "Stop",
		"result", "SUCCESS",
	)
	a.cancel()
}
