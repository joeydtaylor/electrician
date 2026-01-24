package sensor

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// RegisterOnKafkaWriterStart registers callbacks for Kafka writer start.
func (s *Sensor[T]) RegisterOnKafkaWriterStart(callback ...func(types.ComponentMetadata, string, string)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaWriterStart = append(s.OnKafkaWriterStart, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaWriterStart invokes callbacks for Kafka writer start.
func (s *Sensor[T]) InvokeOnKafkaWriterStart(c types.ComponentMetadata, topic, format string) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaWriterStart) {
		if cb == nil {
			continue
		}
		cb(c, topic, format)
	}
}

// RegisterOnKafkaWriterStop registers callbacks for Kafka writer stop.
func (s *Sensor[T]) RegisterOnKafkaWriterStop(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaWriterStop = append(s.OnKafkaWriterStop, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaWriterStop invokes callbacks for Kafka writer stop.
func (s *Sensor[T]) InvokeOnKafkaWriterStop(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaWriterStop) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

// RegisterOnKafkaProduceAttempt registers callbacks for Kafka produce attempts.
func (s *Sensor[T]) RegisterOnKafkaProduceAttempt(callback ...func(types.ComponentMetadata, string, int, int, int)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaProduceAttempt = append(s.OnKafkaProduceAttempt, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaProduceAttempt invokes callbacks for Kafka produce attempts.
func (s *Sensor[T]) InvokeOnKafkaProduceAttempt(c types.ComponentMetadata, topic string, partition, keyBytes, valueBytes int) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaProduceAttempt) {
		if cb == nil {
			continue
		}
		cb(c, topic, partition, keyBytes, valueBytes)
	}
}

// RegisterOnKafkaProduceSuccess registers callbacks for Kafka produce successes.
func (s *Sensor[T]) RegisterOnKafkaProduceSuccess(callback ...func(types.ComponentMetadata, string, int, int64, time.Duration)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaProduceSuccess = append(s.OnKafkaProduceSuccess, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaProduceSuccess invokes callbacks for Kafka produce successes.
func (s *Sensor[T]) InvokeOnKafkaProduceSuccess(c types.ComponentMetadata, topic string, partition int, offset int64, dur time.Duration) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaProduceSuccess) {
		if cb == nil {
			continue
		}
		cb(c, topic, partition, offset, dur)
	}
}

// RegisterOnKafkaProduceError registers callbacks for Kafka produce errors.
func (s *Sensor[T]) RegisterOnKafkaProduceError(callback ...func(types.ComponentMetadata, string, int, error)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaProduceError = append(s.OnKafkaProduceError, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaProduceError invokes callbacks for Kafka produce errors.
func (s *Sensor[T]) InvokeOnKafkaProduceError(c types.ComponentMetadata, topic string, partition int, err error) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaProduceError) {
		if cb == nil {
			continue
		}
		cb(c, topic, partition, err)
	}
}

// RegisterOnKafkaBatchFlush registers callbacks for Kafka batch flush.
func (s *Sensor[T]) RegisterOnKafkaBatchFlush(callback ...func(types.ComponentMetadata, string, int, int, string)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaBatchFlush = append(s.OnKafkaBatchFlush, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaBatchFlush invokes callbacks for Kafka batch flush.
func (s *Sensor[T]) InvokeOnKafkaBatchFlush(c types.ComponentMetadata, topic string, records, bytes int, compression string) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaBatchFlush) {
		if cb == nil {
			continue
		}
		cb(c, topic, records, bytes, compression)
	}
}

// RegisterOnKafkaKeyRendered registers callbacks for Kafka key rendering.
func (s *Sensor[T]) RegisterOnKafkaKeyRendered(callback ...func(types.ComponentMetadata, []byte)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaKeyRendered = append(s.OnKafkaKeyRendered, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaKeyRendered invokes callbacks for Kafka key rendering.
func (s *Sensor[T]) InvokeOnKafkaKeyRendered(c types.ComponentMetadata, key []byte) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaKeyRendered) {
		if cb == nil {
			continue
		}
		cb(c, key)
	}
}

// RegisterOnKafkaHeadersRendered registers callbacks for Kafka header rendering.
func (s *Sensor[T]) RegisterOnKafkaHeadersRendered(callback ...func(types.ComponentMetadata, []struct{ Key, Value string })) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaHeadersRendered = append(s.OnKafkaHeadersRendered, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaHeadersRendered invokes callbacks for Kafka header rendering.
func (s *Sensor[T]) InvokeOnKafkaHeadersRendered(c types.ComponentMetadata, headers []struct{ Key, Value string }) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaHeadersRendered) {
		if cb == nil {
			continue
		}
		cb(c, headers)
	}
}

// RegisterOnKafkaConsumerStart registers callbacks for Kafka consumer start.
func (s *Sensor[T]) RegisterOnKafkaConsumerStart(callback ...func(types.ComponentMetadata, string, []string, string)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaConsumerStart = append(s.OnKafkaConsumerStart, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaConsumerStart invokes callbacks for Kafka consumer start.
func (s *Sensor[T]) InvokeOnKafkaConsumerStart(c types.ComponentMetadata, group string, topics []string, startAt string) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaConsumerStart) {
		if cb == nil {
			continue
		}
		cb(c, group, topics, startAt)
	}
}

// RegisterOnKafkaConsumerStop registers callbacks for Kafka consumer stop.
func (s *Sensor[T]) RegisterOnKafkaConsumerStop(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaConsumerStop = append(s.OnKafkaConsumerStop, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaConsumerStop invokes callbacks for Kafka consumer stop.
func (s *Sensor[T]) InvokeOnKafkaConsumerStop(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaConsumerStop) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

// RegisterOnKafkaPartitionAssigned registers callbacks for partition assignment.
func (s *Sensor[T]) RegisterOnKafkaPartitionAssigned(callback ...func(types.ComponentMetadata, string, int, int64, int64)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaPartitionAssigned = append(s.OnKafkaPartitionAssigned, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaPartitionAssigned invokes callbacks for partition assignment.
func (s *Sensor[T]) InvokeOnKafkaPartitionAssigned(c types.ComponentMetadata, topic string, partition int, startOffset int64, endOffset int64) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaPartitionAssigned) {
		if cb == nil {
			continue
		}
		cb(c, topic, partition, startOffset, endOffset)
	}
}

// RegisterOnKafkaPartitionRevoked registers callbacks for partition revocation.
func (s *Sensor[T]) RegisterOnKafkaPartitionRevoked(callback ...func(types.ComponentMetadata, string, int)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaPartitionRevoked = append(s.OnKafkaPartitionRevoked, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaPartitionRevoked invokes callbacks for partition revocation.
func (s *Sensor[T]) InvokeOnKafkaPartitionRevoked(c types.ComponentMetadata, topic string, partition int) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaPartitionRevoked) {
		if cb == nil {
			continue
		}
		cb(c, topic, partition)
	}
}

// RegisterOnKafkaMessage registers callbacks for consumed messages.
func (s *Sensor[T]) RegisterOnKafkaMessage(callback ...func(types.ComponentMetadata, string, int, int64, int, int)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaMessage = append(s.OnKafkaMessage, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaMessage invokes callbacks for consumed messages.
func (s *Sensor[T]) InvokeOnKafkaMessage(c types.ComponentMetadata, topic string, partition int, offset int64, keyBytes int, valueBytes int) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaMessage) {
		if cb == nil {
			continue
		}
		cb(c, topic, partition, offset, keyBytes, valueBytes)
	}
}

// RegisterOnKafkaDecode registers callbacks for decoded Kafka records.
func (s *Sensor[T]) RegisterOnKafkaDecode(callback ...func(types.ComponentMetadata, string, int, string)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaDecode = append(s.OnKafkaDecode, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaDecode invokes callbacks for decoded Kafka records.
func (s *Sensor[T]) InvokeOnKafkaDecode(c types.ComponentMetadata, topic string, rows int, format string) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaDecode) {
		if cb == nil {
			continue
		}
		cb(c, topic, rows, format)
	}
}

// RegisterOnKafkaCommitSuccess registers callbacks for commit success.
func (s *Sensor[T]) RegisterOnKafkaCommitSuccess(callback ...func(types.ComponentMetadata, string, map[string]int64)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaCommitSuccess = append(s.OnKafkaCommitSuccess, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaCommitSuccess invokes callbacks for commit success.
func (s *Sensor[T]) InvokeOnKafkaCommitSuccess(c types.ComponentMetadata, group string, offsets map[string]int64) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaCommitSuccess) {
		if cb == nil {
			continue
		}
		cb(c, group, offsets)
	}
}

// RegisterOnKafkaCommitError registers callbacks for commit errors.
func (s *Sensor[T]) RegisterOnKafkaCommitError(callback ...func(types.ComponentMetadata, string, error)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaCommitError = append(s.OnKafkaCommitError, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaCommitError invokes callbacks for commit errors.
func (s *Sensor[T]) InvokeOnKafkaCommitError(c types.ComponentMetadata, group string, err error) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaCommitError) {
		if cb == nil {
			continue
		}
		cb(c, group, err)
	}
}

// RegisterOnKafkaDLQProduceAttempt registers callbacks for DLQ produce attempts.
func (s *Sensor[T]) RegisterOnKafkaDLQProduceAttempt(callback ...func(types.ComponentMetadata, string, int, int, int)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaDLQProduceAttempt = append(s.OnKafkaDLQProduceAttempt, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaDLQProduceAttempt invokes callbacks for DLQ produce attempts.
func (s *Sensor[T]) InvokeOnKafkaDLQProduceAttempt(c types.ComponentMetadata, topic string, partition int, keyBytes int, valueBytes int) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaDLQProduceAttempt) {
		if cb == nil {
			continue
		}
		cb(c, topic, partition, keyBytes, valueBytes)
	}
}

// RegisterOnKafkaDLQProduceSuccess registers callbacks for DLQ produce success.
func (s *Sensor[T]) RegisterOnKafkaDLQProduceSuccess(callback ...func(types.ComponentMetadata, string, int, int64, time.Duration)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaDLQProduceSuccess = append(s.OnKafkaDLQProduceSuccess, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaDLQProduceSuccess invokes callbacks for DLQ produce success.
func (s *Sensor[T]) InvokeOnKafkaDLQProduceSuccess(c types.ComponentMetadata, topic string, partition int, offset int64, dur time.Duration) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaDLQProduceSuccess) {
		if cb == nil {
			continue
		}
		cb(c, topic, partition, offset, dur)
	}
}

// RegisterOnKafkaDLQProduceError registers callbacks for DLQ produce errors.
func (s *Sensor[T]) RegisterOnKafkaDLQProduceError(callback ...func(types.ComponentMetadata, string, int, error)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaDLQProduceError = append(s.OnKafkaDLQProduceError, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaDLQProduceError invokes callbacks for DLQ produce errors.
func (s *Sensor[T]) InvokeOnKafkaDLQProduceError(c types.ComponentMetadata, topic string, partition int, err error) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaDLQProduceError) {
		if cb == nil {
			continue
		}
		cb(c, topic, partition, err)
	}
}

// RegisterOnKafkaBillingSample registers callbacks for billing samples.
func (s *Sensor[T]) RegisterOnKafkaBillingSample(callback ...func(types.ComponentMetadata, string, int64, int64)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnKafkaBillingSample = append(s.OnKafkaBillingSample, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnKafkaBillingSample invokes callbacks for billing samples.
func (s *Sensor[T]) InvokeOnKafkaBillingSample(c types.ComponentMetadata, op string, requestUnits int64, bytes int64) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnKafkaBillingSample) {
		if cb == nil {
			continue
		}
		cb(c, op, requestUnits, bytes)
	}
}
