package s3client

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	parquet "github.com/parquet-go/parquet-go"
)

func (a *S3Client[T]) parquetRoller(
	ctx context.Context,
	in <-chan T,
	out chan<- []byte,
	window time.Duration,
	maxRecords int,
	compression parquet.WriterOption,
	compName string,
) error {
	var (
		buf   bytes.Buffer
		pw    *parquet.GenericWriter[T]
		count int
	)

	newWriter := func() *parquet.GenericWriter[T] {
		return parquet.NewGenericWriter[T](&buf, compression)
	}
	pw = newWriter()

	flush := func() error {
		if count == 0 {
			return nil
		}
		if err := pw.Close(); err != nil {
			return err
		}

		for _, sensor := range a.snapshotSensors() {
			if sensor == nil {
				continue
			}
			sensor.InvokeOnS3ParquetRollFlush(a.componentMetadata, count, buf.Len(), compName)
		}

		out <- append([]byte(nil), buf.Bytes()...)
		a.NotifyLoggers(
			types.InfoLevel,
			"Parquet flush",
			"component", a.componentMetadata,
			"event", "ParquetFlush",
			"records", count,
			"bytes", buf.Len(),
			"compression", compName,
		)
		buf.Reset()
		count = 0
		pw = newWriter()
		return nil
	}

	if window <= 0 {
		window = 60 * time.Second
	}
	if maxRecords <= 0 {
		maxRecords = 50_000
	}

	tick := time.NewTicker(window)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = flush()
			close(out)
			return nil
		case <-tick.C:
			if err := flush(); err != nil {
				close(out)
				return err
			}
		case ev, ok := <-in:
			if !ok {
				err := flush()
				close(out)
				return err
			}
			if _, err := pw.Write([]T{ev}); err != nil {
				close(out)
				return err
			}
			count++
			if count >= maxRecords {
				if err := flush(); err != nil {
					close(out)
					return err
				}
			}
		}
	}
}

func (a *S3Client[T]) parquetCompressionFromOpts() parquet.WriterOption {
	val := ""
	if s, ok := a.formatOpts["compression"]; ok {
		val = s
	} else if s, ok := a.formatOpts["parquet_compression"]; ok {
		val = s
	}
	switch strings.ToLower(val) {
	case "zstd":
		return parquet.Compression(&parquet.Zstd)
	case "gzip", "gz":
		return parquet.Compression(&parquet.Gzip)
	default:
		return parquet.Compression(&parquet.Snappy)
	}
}

func (a *S3Client[T]) parquetCompressionNameFromOpts() string {
	if s, ok := a.formatOpts["compression"]; ok {
		return strings.ToLower(s)
	}
	if s, ok := a.formatOpts["parquet_compression"]; ok {
		return strings.ToLower(s)
	}
	return "snappy"
}

func (a *S3Client[T]) startParquetStream(ctx context.Context, in <-chan T) error {
	window := 60 * time.Second
	if s, ok := a.formatOpts["roll_window_ms"]; ok {
		if ms, err := strconv.Atoi(s); err == nil && ms > 0 {
			window = time.Duration(ms) * time.Millisecond
		}
	} else if a.batchMaxAge > 0 {
		window = a.batchMaxAge
	}
	max := 50_000
	if s, ok := a.formatOpts["roll_max_records"]; ok {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			max = n
		}
	} else if a.batchMaxRecords > 0 {
		max = a.batchMaxRecords
	}

	comp := a.parquetCompressionFromOpts()
	compName := a.parquetCompressionNameFromOpts()
	raw := make(chan []byte, 2)

	go func() {
		if err := a.parquetRoller(ctx, in, raw, window, max, comp, compName); err != nil {
			a.NotifyLoggers(
				types.ErrorLevel,
				"parquetRoller failed",
				"component", a.componentMetadata,
				"event", "parquetRoller",
				"error", err,
			)
		}
	}()

	return a.ServeWriterRaw(ctx, raw)
}
