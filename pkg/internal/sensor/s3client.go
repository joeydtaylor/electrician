package sensor

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// RegisterOnS3WriterStart registers callbacks for S3 writer start.
func (s *Sensor[T]) RegisterOnS3WriterStart(callback ...func(types.ComponentMetadata, string, string, string)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3WriterStart = append(s.OnS3WriterStart, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3WriterStart invokes callbacks for S3 writer start.
func (s *Sensor[T]) InvokeOnS3WriterStart(c types.ComponentMetadata, bucket, prefixTpl, format string) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3WriterStart) {
		if cb == nil {
			continue
		}
		cb(c, bucket, prefixTpl, format)
	}
}

// RegisterOnS3WriterStop registers callbacks for S3 writer stop.
func (s *Sensor[T]) RegisterOnS3WriterStop(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3WriterStop = append(s.OnS3WriterStop, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3WriterStop invokes callbacks for S3 writer stop.
func (s *Sensor[T]) InvokeOnS3WriterStop(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3WriterStop) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

// RegisterOnS3KeyRendered registers callbacks for key rendering.
func (s *Sensor[T]) RegisterOnS3KeyRendered(callback ...func(types.ComponentMetadata, string)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3KeyRendered = append(s.OnS3KeyRendered, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3KeyRendered invokes callbacks for key rendering.
func (s *Sensor[T]) InvokeOnS3KeyRendered(c types.ComponentMetadata, key string) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3KeyRendered) {
		if cb == nil {
			continue
		}
		cb(c, key)
	}
}

// RegisterOnS3PutAttempt registers callbacks for PUT attempts.
func (s *Sensor[T]) RegisterOnS3PutAttempt(callback ...func(types.ComponentMetadata, string, string, int, string, string)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3PutAttempt = append(s.OnS3PutAttempt, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3PutAttempt invokes callbacks for PUT attempts.
func (s *Sensor[T]) InvokeOnS3PutAttempt(c types.ComponentMetadata, bucket, key string, bytes int, sseMode, kmsKey string) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3PutAttempt) {
		if cb == nil {
			continue
		}
		cb(c, bucket, key, bytes, sseMode, kmsKey)
	}
}

// RegisterOnS3PutSuccess registers callbacks for PUT success.
func (s *Sensor[T]) RegisterOnS3PutSuccess(callback ...func(types.ComponentMetadata, string, string, int, time.Duration)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3PutSuccess = append(s.OnS3PutSuccess, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3PutSuccess invokes callbacks for PUT success.
func (s *Sensor[T]) InvokeOnS3PutSuccess(c types.ComponentMetadata, bucket, key string, bytes int, dur time.Duration) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3PutSuccess) {
		if cb == nil {
			continue
		}
		cb(c, bucket, key, bytes, dur)
	}
}

// RegisterOnS3PutError registers callbacks for PUT errors.
func (s *Sensor[T]) RegisterOnS3PutError(callback ...func(types.ComponentMetadata, string, string, int, error)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3PutError = append(s.OnS3PutError, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3PutError invokes callbacks for PUT errors.
func (s *Sensor[T]) InvokeOnS3PutError(c types.ComponentMetadata, bucket, key string, bytes int, err error) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3PutError) {
		if cb == nil {
			continue
		}
		cb(c, bucket, key, bytes, err)
	}
}

// RegisterOnS3ParquetRollFlush registers callbacks for parquet roll flush.
func (s *Sensor[T]) RegisterOnS3ParquetRollFlush(callback ...func(types.ComponentMetadata, int, int, string)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3ParquetRollFlush = append(s.OnS3ParquetRollFlush, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3ParquetRollFlush invokes callbacks for parquet roll flush.
func (s *Sensor[T]) InvokeOnS3ParquetRollFlush(c types.ComponentMetadata, records, bytes int, compression string) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3ParquetRollFlush) {
		if cb == nil {
			continue
		}
		cb(c, records, bytes, compression)
	}
}

// RegisterOnS3ReaderListStart registers callbacks for list start.
func (s *Sensor[T]) RegisterOnS3ReaderListStart(callback ...func(types.ComponentMetadata, string, string)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3ReaderListStart = append(s.OnS3ReaderListStart, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3ReaderListStart invokes callbacks for list start.
func (s *Sensor[T]) InvokeOnS3ReaderListStart(c types.ComponentMetadata, bucket, prefix string) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3ReaderListStart) {
		if cb == nil {
			continue
		}
		cb(c, bucket, prefix)
	}
}

// RegisterOnS3ReaderListPage registers callbacks for list pages.
func (s *Sensor[T]) RegisterOnS3ReaderListPage(callback ...func(types.ComponentMetadata, int, bool)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3ReaderListPage = append(s.OnS3ReaderListPage, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3ReaderListPage invokes callbacks for list pages.
func (s *Sensor[T]) InvokeOnS3ReaderListPage(c types.ComponentMetadata, objsInPage int, isTruncated bool) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3ReaderListPage) {
		if cb == nil {
			continue
		}
		cb(c, objsInPage, isTruncated)
	}
}

// RegisterOnS3ReaderObject registers callbacks for reader objects.
func (s *Sensor[T]) RegisterOnS3ReaderObject(callback ...func(types.ComponentMetadata, string, int64)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3ReaderObject = append(s.OnS3ReaderObject, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3ReaderObject invokes callbacks for reader objects.
func (s *Sensor[T]) InvokeOnS3ReaderObject(c types.ComponentMetadata, key string, size int64) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3ReaderObject) {
		if cb == nil {
			continue
		}
		cb(c, key, size)
	}
}

// RegisterOnS3ReaderDecode registers callbacks for decode events.
func (s *Sensor[T]) RegisterOnS3ReaderDecode(callback ...func(types.ComponentMetadata, string, int, string)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3ReaderDecode = append(s.OnS3ReaderDecode, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3ReaderDecode invokes callbacks for decode events.
func (s *Sensor[T]) InvokeOnS3ReaderDecode(c types.ComponentMetadata, key string, rows int, format string) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3ReaderDecode) {
		if cb == nil {
			continue
		}
		cb(c, key, rows, format)
	}
}

// RegisterOnS3ReaderSpillToDisk registers callbacks for spill events.
func (s *Sensor[T]) RegisterOnS3ReaderSpillToDisk(callback ...func(types.ComponentMetadata, int64, int64)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3ReaderSpillToDisk = append(s.OnS3ReaderSpillToDisk, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3ReaderSpillToDisk invokes callbacks for spill events.
func (s *Sensor[T]) InvokeOnS3ReaderSpillToDisk(c types.ComponentMetadata, threshold int64, objectBytes int64) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3ReaderSpillToDisk) {
		if cb == nil {
			continue
		}
		cb(c, threshold, objectBytes)
	}
}

// RegisterOnS3ReaderComplete registers callbacks for reader completion.
func (s *Sensor[T]) RegisterOnS3ReaderComplete(callback ...func(types.ComponentMetadata, int, int)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3ReaderComplete = append(s.OnS3ReaderComplete, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3ReaderComplete invokes callbacks for reader completion.
func (s *Sensor[T]) InvokeOnS3ReaderComplete(c types.ComponentMetadata, objectsScanned int, rowsDecoded int) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3ReaderComplete) {
		if cb == nil {
			continue
		}
		cb(c, objectsScanned, rowsDecoded)
	}
}

// RegisterOnS3BillingSample registers callbacks for billing samples.
func (s *Sensor[T]) RegisterOnS3BillingSample(callback ...func(types.ComponentMetadata, string, int64, int64, string)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnS3BillingSample = append(s.OnS3BillingSample, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnS3BillingSample invokes callbacks for billing samples.
func (s *Sensor[T]) InvokeOnS3BillingSample(c types.ComponentMetadata, op string, requestUnits int64, bytes int64, storageClass string) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnS3BillingSample) {
		if cb == nil {
			continue
		}
		cb(c, op, requestUnits, bytes, storageClass)
	}
}
