// pkg/internal/types/codec.go
package types

import "io"

////////////////////////////////////////////////////////////////////////////////
// Existing interfaces (kept for back-compat)
////////////////////////////////////////////////////////////////////////////////

// Decodable is an optional helper for types that can self-decode.
type Decodable interface {
	DecodeJSON(data io.Reader) error
	DecodeXML(data io.Reader) error
	DecodeText(data string) error
	DecodeBinary(data []byte) error
}

// Decoder deserializes one object from r.
type Decoder[T any] interface {
	Decode(io.Reader) (T, error)
}

// Encoder serializes one object to w.
type Encoder[T any] interface {
	Encode(io.Writer, T) error
}

// JSONDecoder decodes one or many JSON objects.
type JSONDecoder[T any] interface {
	Decode(io.Reader) (T, error)
	DecodeSlice(r io.Reader) ([]T, error)
}

// JSONEncoder encodes one or many JSON objects.
type JSONEncoder[T any] interface {
	Encode(io.Writer, T) error
	EncodeSlice(w io.Writer, elems []T) error
}

////////////////////////////////////////////////////////////////////////////////
// Extensions for streaming/batched formats (NDJSON, Parquet, etc.)
////////////////////////////////////////////////////////////////////////////////

// FormatWriterOptions/FormatReaderOptions carry per-format knobs.
type FormatWriterOptions struct {
	// e.g. schema version/hash for T
	SchemaVersion string
	// free-form: {"gzip":"true"}, {"compression":"zstd","row_group_bytes":"134217728"}, etc.
	Extra map[string]string
}

type FormatReaderOptions struct {
	// optional projection for columnar formats
	Columns []string
	// free-form: predicate hints, etc.
	Extra map[string]string
}

// RecordWriter supports streaming write of records (row-at-a-time) to an underlying io.Writer.
type RecordWriter[T any] interface {
	// Initialize writer with the destination sink and options.
	Begin(w io.Writer, opts FormatWriterOptions) error
	// Write one record (buffering allowed internally).
	Write(rec T) error
	// Flush internal buffers (optional for some formats).
	Flush() error
	// Close and finalize the stream (must release resources).
	Close() error
	// ContentType/Ext provide S3 metadata and filename extension for the produced stream.
	ContentType() string // e.g. "application/x-ndjson", "application/parquet"
	Ext() string         // e.g. ".ndjson", ".parquet"
}

// RecordReader supports streaming read of records from an underlying io.Reader.
type RecordReader[T any] interface {
	// Initialize reader with source and options.
	Begin(r io.Reader, opts FormatReaderOptions) error
	// Iterator-style decode.
	Next() bool
	Record() T
	Err() error
	// Close and release resources.
	Close() error
}

// Format binds a name to concrete reader/writer implementations for type T.
type Format[T any] interface {
	// Canonical name: "ndjson", "parquet", "raw"
	Name() string
	// Factory methods. Implementations may capture per-format defaults.
	NewRecordWriter() RecordWriter[T]
	NewRecordReader() RecordReader[T]
}

////////////////////////////////////////////////////////////////////////////////
// Adapters for back-compat (optional usage)
// These let existing code that expects Encoder/Decoder work over a RecordWriter/Reader.
////////////////////////////////////////////////////////////////////////////////

type encoderAdapter[T any] struct {
	wr   RecordWriter[T]
	opts FormatWriterOptions
	init bool
}

func NewEncoderFromRecordWriter[T any](wr RecordWriter[T], opts FormatWriterOptions, sink io.Writer) Encoder[T] {
	return &encoderAdapter[T]{wr: wr, opts: opts, init: initBegin(wr, sink, opts)}
}

func (a *encoderAdapter[T]) Encode(w io.Writer, v T) error {
	// If sink changes between calls, re-Begin.
	if !a.init {
		if err := a.wr.Begin(w, a.opts); err != nil {
			return err
		}
		a.init = true
	}
	if err := a.wr.Write(v); err != nil {
		return err
	}
	// No implicit Close; caller controls lifecycle.
	return nil
}

type decoderAdapter[T any] struct {
	rd   RecordReader[T]
	opts FormatReaderOptions
	init bool
	src  io.Reader
}

func NewDecoderFromRecordReader[T any](rd RecordReader[T], opts FormatReaderOptions, src io.Reader) Decoder[T] {
	_ = initBeginReader(rd, src, opts) // eager init; ignore here, real call in Decode
	return &decoderAdapter[T]{rd: rd, opts: opts, src: src}
}

func (a *decoderAdapter[T]) Decode(r io.Reader) (T, error) {
	var zero T
	if !a.init || r != a.src {
		if err := a.rd.Begin(r, a.opts); err != nil {
			return zero, err
		}
		a.init = true
		a.src = r
	}
	if !a.rd.Next() {
		if err := a.rd.Err(); err != nil {
			return zero, err
		}
		return zero, io.EOF
	}
	return a.rd.Record(), nil
}

////////////////////////////////////////////////////////////////////////////////
// tiny helpers
////////////////////////////////////////////////////////////////////////////////

func initBegin[T any](w RecordWriter[T], sink io.Writer, opts FormatWriterOptions) bool {
	if err := w.Begin(sink, opts); err != nil {
		return false
	}
	return true
}

func initBeginReader[T any](r RecordReader[T], src io.Reader, opts FormatReaderOptions) bool {
	if err := r.Begin(src, opts); err != nil {
		return false
	}
	return true
}
