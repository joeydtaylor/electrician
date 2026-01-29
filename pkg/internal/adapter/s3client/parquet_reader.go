package s3client

import (
	"bytes"
	"io"
	"os"
	"strconv"

	s3api "github.com/aws/aws-sdk-go-v2/service/s3"
	parquet "github.com/parquet-go/parquet-go"
)

func (a *S3Client[T]) parquetRowsFromBody(get *s3api.GetObjectOutput) ([]T, error) {
	const def int64 = 64 << 20
	thr := def
	if s, ok := a.readerFormatOpts["spill_threshold_bytes"]; ok {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil && v > 0 {
			thr = v
		}
	}

	if get.ContentLength != nil && *get.ContentLength > thr {
		f, err := os.CreateTemp("", "electrician-parquet-*")
		if err != nil {
			_ = get.Body.Close()
			return nil, err
		}
		defer func() {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}()
		if _, err := io.Copy(f, get.Body); err != nil {
			_ = get.Body.Close()
			return nil, err
		}
		_ = get.Body.Close()

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		return a.readParquetAllFromReaderAt(f)
	}

	data, err := io.ReadAll(get.Body)
	_ = get.Body.Close()
	if err != nil {
		return nil, err
	}
	return a.readParquetAllFromReaderAt(bytes.NewReader(data))
}

func (a *S3Client[T]) parquetRowsFromBytes(data []byte) ([]T, error) {
	return a.readParquetAllFromReaderAt(bytes.NewReader(data))
}

func (a *S3Client[T]) readParquetAllFromReaderAt(ra io.ReaderAt) ([]T, error) {
	gr := parquet.NewGenericReader[T](ra)
	defer gr.Close()

	out := make([]T, 0, 1024)
	batch := make([]T, 1024)
	for {
		n, err := gr.Read(batch)
		if n > 0 {
			out = append(out, batch[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
