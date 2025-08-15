// pkg/internal/sensor/kafka_hooks.go
package sensor

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ---------- Writer lifecycle ----------

func (m *Sensor[T]) RegisterOnKafkaWriterStart(callback ...func(types.ComponentMetadata, string, string)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaWriterStart = append(m.OnKafkaWriterStart, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaWriterStart, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaWriterStart(c types.ComponentMetadata, topic, format string) {
	for _, cb := range m.OnKafkaWriterStart {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaWriterStart, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, topic, format)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaWriterStop(callback ...func(types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaWriterStop = append(m.OnKafkaWriterStop, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaWriterStop, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaWriterStop(c types.ComponentMetadata) {
	for _, cb := range m.OnKafkaWriterStop {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaWriterStop, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c)
		m.callbackLock.Unlock()
	}
}

// ---------- Produce lifecycle ----------

func (m *Sensor[T]) RegisterOnKafkaProduceAttempt(callback ...func(types.ComponentMetadata, string, int, int, int)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaProduceAttempt = append(m.OnKafkaProduceAttempt, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaProduceAttempt, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaProduceAttempt(c types.ComponentMetadata, topic string, partition, keyBytes, valueBytes int) {
	for _, cb := range m.OnKafkaProduceAttempt {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaProduceAttempt, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, topic, partition, keyBytes, valueBytes)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaProduceSuccess(callback ...func(types.ComponentMetadata, string, int, int64, time.Duration)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaProduceSuccess = append(m.OnKafkaProduceSuccess, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaProduceSuccess, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaProduceSuccess(c types.ComponentMetadata, topic string, partition int, offset int64, dur time.Duration) {
	for _, cb := range m.OnKafkaProduceSuccess {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaProduceSuccess, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, topic, partition, offset, dur)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaProduceError(callback ...func(types.ComponentMetadata, string, int, error)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaProduceError = append(m.OnKafkaProduceError, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaProduceError, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaProduceError(c types.ComponentMetadata, topic string, partition int, err error) {
	for _, cb := range m.OnKafkaProduceError {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaProduceError, error: %v, target: %v", m.componentMetadata, err, cb)
		m.callbackLock.Lock()
		cb(c, topic, partition, err)
		m.callbackLock.Unlock()
	}
}

// ---------- Writer batching/flush ----------

func (m *Sensor[T]) RegisterOnKafkaBatchFlush(callback ...func(types.ComponentMetadata, string, int, int, string)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaBatchFlush = append(m.OnKafkaBatchFlush, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaBatchFlush, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaBatchFlush(c types.ComponentMetadata, topic string, records, bytes int, compression string) {
	for _, cb := range m.OnKafkaBatchFlush {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaBatchFlush, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, topic, records, bytes, compression)
		m.callbackLock.Unlock()
	}
}

// ---------- Record adornments ----------

func (m *Sensor[T]) RegisterOnKafkaKeyRendered(callback ...func(types.ComponentMetadata, []byte)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaKeyRendered = append(m.OnKafkaKeyRendered, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaKeyRendered, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaKeyRendered(c types.ComponentMetadata, key []byte) {
	for _, cb := range m.OnKafkaKeyRendered {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaKeyRendered, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, key)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaHeadersRendered(callback ...func(types.ComponentMetadata, []struct{ Key, Value string })) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaHeadersRendered = append(m.OnKafkaHeadersRendered, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaHeadersRendered, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaHeadersRendered(c types.ComponentMetadata, headers []struct{ Key, Value string }) {
	for _, cb := range m.OnKafkaHeadersRendered {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaHeadersRendered, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, headers)
		m.callbackLock.Unlock()
	}
}

// ---------- Consumer lifecycle & flow ----------

func (m *Sensor[T]) RegisterOnKafkaConsumerStart(callback ...func(types.ComponentMetadata, string, []string, string)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaConsumerStart = append(m.OnKafkaConsumerStart, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaConsumerStart, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaConsumerStart(c types.ComponentMetadata, group string, topics []string, startAt string) {
	for _, cb := range m.OnKafkaConsumerStart {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaConsumerStart, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, group, topics, startAt)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaConsumerStop(callback ...func(types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaConsumerStop = append(m.OnKafkaConsumerStop, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaConsumerStop, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaConsumerStop(c types.ComponentMetadata) {
	for _, cb := range m.OnKafkaConsumerStop {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaConsumerStop, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaPartitionAssigned(callback ...func(types.ComponentMetadata, string, int, int64, int64)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaPartitionAssigned = append(m.OnKafkaPartitionAssigned, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaPartitionAssigned, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaPartitionAssigned(c types.ComponentMetadata, topic string, partition int, startOffset, endOffset int64) {
	for _, cb := range m.OnKafkaPartitionAssigned {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaPartitionAssigned, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, topic, partition, startOffset, endOffset)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaPartitionRevoked(callback ...func(types.ComponentMetadata, string, int)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaPartitionRevoked = append(m.OnKafkaPartitionRevoked, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaPartitionRevoked, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaPartitionRevoked(c types.ComponentMetadata, topic string, partition int) {
	for _, cb := range m.OnKafkaPartitionRevoked {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaPartitionRevoked, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, topic, partition)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaMessage(callback ...func(types.ComponentMetadata, string, int, int64, int, int)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaMessage = append(m.OnKafkaMessage, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaMessage, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaMessage(c types.ComponentMetadata, topic string, partition int, offset int64, keyBytes, valueBytes int) {
	for _, cb := range m.OnKafkaMessage {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaMessage, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, topic, partition, offset, keyBytes, valueBytes)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaDecode(callback ...func(types.ComponentMetadata, string, int, string)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaDecode = append(m.OnKafkaDecode, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaDecode, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaDecode(c types.ComponentMetadata, topic string, rows int, format string) {
	for _, cb := range m.OnKafkaDecode {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaDecode, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, topic, rows, format)
		m.callbackLock.Unlock()
	}
}

// ---------- Commits ----------

func (m *Sensor[T]) RegisterOnKafkaCommitSuccess(callback ...func(types.ComponentMetadata, string, map[string]int64)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaCommitSuccess = append(m.OnKafkaCommitSuccess, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaCommitSuccess, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaCommitSuccess(c types.ComponentMetadata, group string, offsets map[string]int64) {
	for _, cb := range m.OnKafkaCommitSuccess {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaCommitSuccess, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, group, offsets)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaCommitError(callback ...func(types.ComponentMetadata, string, error)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaCommitError = append(m.OnKafkaCommitError, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaCommitError, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaCommitError(c types.ComponentMetadata, group string, err error) {
	for _, cb := range m.OnKafkaCommitError {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaCommitError, error: %v, target: %v", m.componentMetadata, err, cb)
		m.callbackLock.Lock()
		cb(c, group, err)
		m.callbackLock.Unlock()
	}
}

// ---------- DLQ ----------

func (m *Sensor[T]) RegisterOnKafkaDLQProduceAttempt(callback ...func(types.ComponentMetadata, string, int, int, int)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaDLQProduceAttempt = append(m.OnKafkaDLQProduceAttempt, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaDLQProduceAttempt, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaDLQProduceAttempt(c types.ComponentMetadata, dlqTopic string, partition, keyBytes, valueBytes int) {
	for _, cb := range m.OnKafkaDLQProduceAttempt {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaDLQProduceAttempt, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, dlqTopic, partition, keyBytes, valueBytes)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaDLQProduceSuccess(callback ...func(types.ComponentMetadata, string, int, int64, time.Duration)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaDLQProduceSuccess = append(m.OnKafkaDLQProduceSuccess, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaDLQProduceSuccess, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaDLQProduceSuccess(c types.ComponentMetadata, dlqTopic string, partition int, offset int64, dur time.Duration) {
	for _, cb := range m.OnKafkaDLQProduceSuccess {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaDLQProduceSuccess, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, dlqTopic, partition, offset, dur)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnKafkaDLQProduceError(callback ...func(types.ComponentMetadata, string, int, error)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaDLQProduceError = append(m.OnKafkaDLQProduceError, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaDLQProduceError, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaDLQProduceError(c types.ComponentMetadata, dlqTopic string, partition int, err error) {
	for _, cb := range m.OnKafkaDLQProduceError {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaDLQProduceError, error: %v, target: %v", m.componentMetadata, err, cb)
		m.callbackLock.Lock()
		cb(c, dlqTopic, partition, err)
		m.callbackLock.Unlock()
	}
}

// ---------- Billing (optional) ----------

func (m *Sensor[T]) RegisterOnKafkaBillingSample(callback ...func(types.ComponentMetadata, string, int64, int64)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnKafkaBillingSample = append(m.OnKafkaBillingSample, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: RegisterOnKafkaBillingSample, target: %v", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnKafkaBillingSample(c types.ComponentMetadata, op string, requestUnits, bytes int64) {
	for _, cb := range m.OnKafkaBillingSample {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: InvokeOnKafkaBillingSample, target: %v", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, op, requestUnits, bytes)
		m.callbackLock.Unlock()
	}
}
