// pkg/internal/sensor/s3_hooks.go
package sensor

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ---------- Writer lifecycle ----------

func (m *Sensor[T]) RegisterOnS3WriterStart(callback ...func(types.ComponentMetadata, string, string, string)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3WriterStart = append(m.OnS3WriterStart, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3WriterStart, target: %v => Registered OnS3WriterStart function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3WriterStart(c types.ComponentMetadata, bucket, prefixTpl, format string) {
	for _, cb := range m.OnS3WriterStart {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3WriterStart, target: %v => Invoked OnS3WriterStart function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, bucket, prefixTpl, format)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnS3WriterStop(callback ...func(types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3WriterStop = append(m.OnS3WriterStop, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3WriterStop, target: %v => Registered OnS3WriterStop function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3WriterStop(c types.ComponentMetadata) {
	for _, cb := range m.OnS3WriterStop {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3WriterStop, target: %v => Invoked OnS3WriterStop function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c)
		m.callbackLock.Unlock()
	}
}

// ---------- Key/PUT lifecycle ----------

func (m *Sensor[T]) RegisterOnS3KeyRendered(callback ...func(types.ComponentMetadata, string)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3KeyRendered = append(m.OnS3KeyRendered, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3KeyRendered, target: %v => Registered OnS3KeyRendered function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3KeyRendered(c types.ComponentMetadata, key string) {
	for _, cb := range m.OnS3KeyRendered {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3KeyRendered, target: %v => Invoked OnS3KeyRendered function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, key)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnS3PutAttempt(callback ...func(types.ComponentMetadata, string, string, int, string, string)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3PutAttempt = append(m.OnS3PutAttempt, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3PutAttempt, target: %v => Registered OnS3PutAttempt function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3PutAttempt(c types.ComponentMetadata, bucket, key string, bytes int, sseMode, kmsKey string) {
	for _, cb := range m.OnS3PutAttempt {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3PutAttempt, target: %v => Invoked OnS3PutAttempt function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, bucket, key, bytes, sseMode, kmsKey)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnS3PutSuccess(callback ...func(types.ComponentMetadata, string, string, int, time.Duration)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3PutSuccess = append(m.OnS3PutSuccess, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3PutSuccess, target: %v => Registered OnS3PutSuccess function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3PutSuccess(c types.ComponentMetadata, bucket, key string, bytes int, dur time.Duration) {
	for _, cb := range m.OnS3PutSuccess {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3PutSuccess, target: %v => Invoked OnS3PutSuccess function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, bucket, key, bytes, dur)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnS3PutError(callback ...func(types.ComponentMetadata, string, string, int, error)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3PutError = append(m.OnS3PutError, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3PutError, target: %v => Registered OnS3PutError function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3PutError(c types.ComponentMetadata, bucket, key string, bytes int, err error) {
	for _, cb := range m.OnS3PutError {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3PutError, error: %v, target: %v => Invoked OnS3PutError function", m.componentMetadata, err, cb)
		m.callbackLock.Lock()
		cb(c, bucket, key, bytes, err)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnS3ParquetRollFlush(callback ...func(types.ComponentMetadata, int, int, string)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3ParquetRollFlush = append(m.OnS3ParquetRollFlush, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3ParquetRollFlush, target: %v => Registered OnS3ParquetRollFlush function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3ParquetRollFlush(c types.ComponentMetadata, records, bytes int, compression string) {
	for _, cb := range m.OnS3ParquetRollFlush {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3ParquetRollFlush, target: %v => Invoked OnS3ParquetRollFlush function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, records, bytes, compression)
		m.callbackLock.Unlock()
	}
}

// ---------- Reader lifecycle ----------

func (m *Sensor[T]) RegisterOnS3ReaderListStart(callback ...func(types.ComponentMetadata, string, string)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3ReaderListStart = append(m.OnS3ReaderListStart, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3ReaderListStart, target: %v => Registered OnS3ReaderListStart function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3ReaderListStart(c types.ComponentMetadata, bucket, prefix string) {
	for _, cb := range m.OnS3ReaderListStart {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3ReaderListStart, target: %v => Invoked OnS3ReaderListStart function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, bucket, prefix)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnS3ReaderListPage(callback ...func(types.ComponentMetadata, int, bool)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3ReaderListPage = append(m.OnS3ReaderListPage, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3ReaderListPage, target: %v => Registered OnS3ReaderListPage function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3ReaderListPage(c types.ComponentMetadata, objsInPage int, isTruncated bool) {
	for _, cb := range m.OnS3ReaderListPage {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3ReaderListPage, target: %v => Invoked OnS3ReaderListPage function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, objsInPage, isTruncated)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnS3ReaderObject(callback ...func(types.ComponentMetadata, string, int64)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3ReaderObject = append(m.OnS3ReaderObject, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3ReaderObject, target: %v => Registered OnS3ReaderObject function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3ReaderObject(c types.ComponentMetadata, key string, size int64) {
	for _, cb := range m.OnS3ReaderObject {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3ReaderObject, target: %v => Invoked OnS3ReaderObject function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, key, size)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnS3ReaderDecode(callback ...func(types.ComponentMetadata, string, int, string)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3ReaderDecode = append(m.OnS3ReaderDecode, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3ReaderDecode, target: %v => Registered OnS3ReaderDecode function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3ReaderDecode(c types.ComponentMetadata, key string, rows int, format string) {
	for _, cb := range m.OnS3ReaderDecode {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3ReaderDecode, target: %v => Invoked OnS3ReaderDecode function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, key, rows, format)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnS3ReaderSpillToDisk(callback ...func(types.ComponentMetadata, int64, int64)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3ReaderSpillToDisk = append(m.OnS3ReaderSpillToDisk, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3ReaderSpillToDisk, target: %v => Registered OnS3ReaderSpillToDisk function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3ReaderSpillToDisk(c types.ComponentMetadata, threshold, objectBytes int64) {
	for _, cb := range m.OnS3ReaderSpillToDisk {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3ReaderSpillToDisk, target: %v => Invoked OnS3ReaderSpillToDisk function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, threshold, objectBytes)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) RegisterOnS3ReaderComplete(callback ...func(types.ComponentMetadata, int, int)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3ReaderComplete = append(m.OnS3ReaderComplete, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3ReaderComplete, target: %v => Registered OnS3ReaderComplete function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3ReaderComplete(c types.ComponentMetadata, objectsScanned, rowsDecoded int) {
	for _, cb := range m.OnS3ReaderComplete {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3ReaderComplete, target: %v => Invoked OnS3ReaderComplete function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, objectsScanned, rowsDecoded)
		m.callbackLock.Unlock()
	}
}

// ---------- Billing sample (optional) ----------

func (m *Sensor[T]) RegisterOnS3BillingSample(callback ...func(types.ComponentMetadata, string, int64, int64, string)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnS3BillingSample = append(m.OnS3BillingSample, callback...)
	for _, cb := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnS3BillingSample, target: %v => Registered OnS3BillingSample function", m.componentMetadata, cb)
	}
}

func (m *Sensor[T]) InvokeOnS3BillingSample(c types.ComponentMetadata, op string, requestUnits, bytes int64, storageClass string) {
	for _, cb := range m.OnS3BillingSample {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnS3BillingSample, target: %v => Invoked OnS3BillingSample function", m.componentMetadata, cb)
		m.callbackLock.Lock()
		cb(c, op, requestUnits, bytes, storageClass)
		m.callbackLock.Unlock()
	}
}
