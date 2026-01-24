package s3client

// Stop terminates writer activity and emits stop hooks.
func (a *S3Client[T]) Stop() {
	for _, sensor := range a.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnS3WriterStop(a.componentMetadata)
	}
	a.cancel()
}
