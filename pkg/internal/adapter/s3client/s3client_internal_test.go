package s3client

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	parquet "github.com/parquet-go/parquet-go"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newTestS3Client(rt http.RoundTripper) *s3.Client {
	cfg := aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		HTTPClient:  &http.Client{Transport: rt},
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.EndpointResolver = s3.EndpointResolverFromURL("https://s3.test")
	})
}

func keyFromPath(p string) string {
	trimmed := strings.TrimPrefix(p, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func listObjectsXML(keys []string, truncated bool, token string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString(`<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">`)
	b.WriteString(fmt.Sprintf("<IsTruncated>%t</IsTruncated>", truncated))
	if token != "" {
		b.WriteString("<NextContinuationToken>" + token + "</NextContinuationToken>")
	}
	for _, k := range keys {
		b.WriteString("<Contents><Key>" + k + "</Key><Size>1</Size></Contents>")
	}
	b.WriteString(`</ListBucketResult>`)
	return b.String()
}

func httpResponse(status int, body []byte, headers map[string]string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     h,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

type sampleRow struct {
	ID   int64  `json:"id" parquet:"name=id,type=INT64"`
	Name string `json:"name" parquet:"name=name,type=BYTE_ARRAY,convertedtype=UTF8"`
}

func TestS3Client_FetchNDJSON(t *testing.T) {
	plain := []byte(`{"id":1,"name":"a"}` + "\n" + `{"id":2,"name":"b"}` + "\n")
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, _ = zw.Write([]byte(`{"id":3,"name":"c"}` + "\n"))
	_ = zw.Close()

	keys := []string{"prefix/one.ndjson", "prefix/two.ndjson"}
	bodies := map[string]struct {
		data    []byte
		encoded bool
	}{
		keys[0]: {data: plain, encoded: false},
		keys[1]: {data: gz.Bytes(), encoded: true},
	}

	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodGet && r.URL.Query().Get("list-type") == "2" {
			return httpResponse(http.StatusOK, []byte(listObjectsXML(keys, false, "")), map[string]string{
				"Content-Type": "application/xml",
			}), nil
		}
		if r.Method == http.MethodGet {
			key := keyFromPath(r.URL.Path)
			body, ok := bodies[key]
			if !ok {
				return nil, fmt.Errorf("unexpected key %s", key)
			}
			headers := map[string]string{
				"Content-Type": "application/octet-stream",
			}
			if body.encoded {
				headers["Content-Encoding"] = "gzip"
			}
			return httpResponse(http.StatusOK, body.data, headers), nil
		}
		return nil, fmt.Errorf("unexpected method %s", r.Method)
	})

	client := NewS3ClientAdapter[sampleRow](context.Background()).(*S3Client[sampleRow])
	client.cli = newTestS3Client(rt)
	client.bucket = "bucket"
	client.listPrefix = "prefix/"

	resp, err := client.Fetch()
	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}
	if resp.StatusCode != 200 || len(resp.Body) != 3 {
		t.Fatalf("expected 3 rows, got status=%d len=%d", resp.StatusCode, len(resp.Body))
	}
	if client.listStartAfter != keys[1] {
		t.Fatalf("expected listStartAfter to be %q, got %q", keys[1], client.listStartAfter)
	}
}

func TestS3Client_FetchTruncated(t *testing.T) {
	keys1 := []string{"prefix/one.ndjson"}
	keys2 := []string{"prefix/two.ndjson"}

	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodGet && r.URL.Query().Get("list-type") == "2" {
			token := r.URL.Query().Get("continuation-token")
			if token == "" {
				return httpResponse(http.StatusOK, []byte(listObjectsXML(keys1, true, "next")), map[string]string{
					"Content-Type": "application/xml",
				}), nil
			}
			if token == "next" {
				return httpResponse(http.StatusOK, []byte(listObjectsXML(keys2, false, "")), map[string]string{
					"Content-Type": "application/xml",
				}), nil
			}
		}
		if r.Method == http.MethodGet {
			key := keyFromPath(r.URL.Path)
			body := []byte(`{"id":1,"name":"a"}` + "\n")
			if key == keys2[0] {
				body = []byte(`{"id":2,"name":"b"}` + "\n")
			}
			return httpResponse(http.StatusOK, body, map[string]string{
				"Content-Type": "application/octet-stream",
			}), nil
		}
		return nil, fmt.Errorf("unexpected request")
	})

	client := NewS3ClientAdapter[sampleRow](context.Background()).(*S3Client[sampleRow])
	client.cli = newTestS3Client(rt)
	client.bucket = "bucket"
	client.listPrefix = "prefix/"

	resp, err := client.Fetch()
	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}
	if resp.StatusCode != 200 || len(resp.Body) != 2 {
		t.Fatalf("expected 2 rows, got status=%d len=%d", resp.StatusCode, len(resp.Body))
	}
	if client.listStartAfter != keys2[0] {
		t.Fatalf("expected listStartAfter %q, got %q", keys2[0], client.listStartAfter)
	}
}

func TestS3Client_FetchNoResults(t *testing.T) {
	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodGet && r.URL.Query().Get("list-type") == "2" {
			return httpResponse(http.StatusOK, []byte(listObjectsXML(nil, false, "")), map[string]string{
				"Content-Type": "application/xml",
			}), nil
		}
		return nil, fmt.Errorf("unexpected request")
	})

	client := NewS3ClientAdapter[sampleRow](context.Background()).(*S3Client[sampleRow])
	client.cli = newTestS3Client(rt)
	client.bucket = "bucket"

	resp, err := client.Fetch()
	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}
	if resp.StatusCode != 204 {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestS3Client_ServeOneShot(t *testing.T) {
	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == http.MethodGet && r.URL.Query().Get("list-type") == "2" {
			return httpResponse(http.StatusOK, []byte(listObjectsXML([]string{"prefix/one.ndjson"}, false, "")), map[string]string{
				"Content-Type": "application/xml",
			}), nil
		}
		if r.Method == http.MethodGet {
			return httpResponse(http.StatusOK, []byte(`{"id":1,"name":"a"}`+"\n"), map[string]string{
				"Content-Type": "application/octet-stream",
			}), nil
		}
		return nil, fmt.Errorf("unexpected request")
	})

	client := NewS3ClientAdapter[sampleRow](context.Background()).(*S3Client[sampleRow])
	client.cli = newTestS3Client(rt)
	client.bucket = "bucket"
	client.listPollInterval = 0

	var got int32
	if err := client.Serve(context.Background(), func(_ context.Context, v sampleRow) error {
		atomic.AddInt32(&got, 1)
		if v.ID != 1 {
			t.Fatalf("unexpected row: %+v", v)
		}
		return nil
	}); err != nil {
		t.Fatalf("Serve error: %v", err)
	}
	if atomic.LoadInt32(&got) != 1 {
		t.Fatalf("expected 1 submission, got %d", got)
	}
}

func TestS3Client_PutWithRetry(t *testing.T) {
	var calls int32
	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPut {
			return nil, fmt.Errorf("unexpected method %s", r.Method)
		}
		count := atomic.AddInt32(&calls, 1)
		if count == 1 {
			return httpResponse(http.StatusServiceUnavailable, []byte(`<Error>Service Unavailable</Error>`), map[string]string{
				"Content-Type": "application/xml",
			}), nil
		}
		return httpResponse(http.StatusOK, nil, map[string]string{}), nil
	})

	client := NewS3ClientAdapter[sampleRow](context.Background()).(*S3Client[sampleRow])
	client.cli = newTestS3Client(rt)
	client.bucket = "bucket"

	payload := bytes.NewReader([]byte("data"))
	put := &s3.PutObjectInput{
		Bucket: &client.bucket,
		Key:    aws.String("k"),
		Body:   payload,
	}

	if _, err := client.putWithRetry(context.Background(), put, "k", 4); err != nil {
		t.Fatalf("putWithRetry error: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 attempts, got %d", calls)
	}
}

func TestS3Client_PutWithRetryNonRetryable(t *testing.T) {
	var calls int32
	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return httpResponse(http.StatusBadRequest, []byte(`<Error>bad</Error>`), map[string]string{
			"Content-Type": "application/xml",
		}), nil
	})

	client := NewS3ClientAdapter[sampleRow](context.Background()).(*S3Client[sampleRow])
	client.cli = newTestS3Client(rt)
	client.bucket = "bucket"

	payload := bytes.NewReader([]byte("data"))
	put := &s3.PutObjectInput{
		Bucket: &client.bucket,
		Key:    aws.String("k"),
		Body:   payload,
	}

	if _, err := client.putWithRetry(context.Background(), put, "k", 4); err == nil {
		t.Fatalf("expected error")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 attempt, got %d", calls)
	}
}

func TestS3Client_FlushNDJSONGzip(t *testing.T) {
	var got struct {
		key     string
		body    []byte
		encoded string
		ct      string
	}

	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPut {
			return nil, fmt.Errorf("unexpected method %s", r.Method)
		}
		got.key = keyFromPath(r.URL.Path)
		got.encoded = r.Header.Get("Content-Encoding")
		got.ct = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		got.body = body
		return httpResponse(http.StatusOK, nil, map[string]string{}), nil
	})

	client := NewS3ClientAdapter[sampleRow](context.Background()).(*S3Client[sampleRow])
	client.cli = newTestS3Client(rt)
	client.bucket = "bucket"
	client.prefixTemplate = "prefix"
	client.fileNameTmpl = "file"
	client.ndjsonEncGz = true
	client.formatOpts = map[string]string{"gzip": "true"}

	_ = client.writeOne(sampleRow{ID: 1, Name: "a"})
	_ = client.writeOne(sampleRow{ID: 2, Name: "b"})
	if err := client.flush(context.Background(), time.Now()); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	if !strings.HasSuffix(got.key, ".ndjson") {
		t.Fatalf("expected .ndjson key, got %q", got.key)
	}
	if got.encoded != "gzip" {
		t.Fatalf("expected gzip encoding, got %q", got.encoded)
	}
	if got.ct == "" {
		t.Fatalf("expected content-type to be set")
	}

	zr, err := gzip.NewReader(bytes.NewReader(got.body))
	if err != nil {
		t.Fatalf("gzip reader error: %v", err)
	}
	decompressed, _ := io.ReadAll(zr)
	_ = zr.Close()
	if !strings.Contains(string(decompressed), `"id":1`) {
		t.Fatalf("expected decompressed body to include record")
	}
}

func TestS3Client_FlushRawDefaults(t *testing.T) {
	var got struct {
		key string
		ct  string
	}

	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		got.key = keyFromPath(r.URL.Path)
		got.ct = r.Header.Get("Content-Type")
		return httpResponse(http.StatusOK, nil, map[string]string{}), nil
	})

	client := NewS3ClientAdapter[sampleRow](context.Background()).(*S3Client[sampleRow])
	client.cli = newTestS3Client(rt)
	client.bucket = "bucket"
	client.prefixTemplate = "prefix"
	client.fileNameTmpl = "file"
	client.formatName = "parquet"
	client.rawWriterExt = ""
	client.rawWriterContentType = ""

	client.buf.Reset()
	_, _ = client.buf.Write([]byte("blob"))
	client.count = 1

	if err := client.flushRaw(context.Background(), time.Now()); err != nil {
		t.Fatalf("flushRaw error: %v", err)
	}

	if !strings.HasSuffix(got.key, ".parquet") {
		t.Fatalf("expected parquet key, got %q", got.key)
	}
	if got.ct != "application/parquet" {
		t.Fatalf("expected application/parquet content-type, got %q", got.ct)
	}
}

func TestS3Client_ParquetRowsFromBody(t *testing.T) {
	var buf bytes.Buffer
	pw := parquet.NewGenericWriter[sampleRow](&buf)
	_, err := pw.Write([]sampleRow{{ID: 1, Name: "a"}})
	if err != nil {
		t.Fatalf("parquet write error: %v", err)
	}
	if err := pw.Close(); err != nil {
		t.Fatalf("parquet close error: %v", err)
	}

	client := NewS3ClientAdapter[sampleRow](context.Background()).(*S3Client[sampleRow])
	out, err := client.parquetRowsFromBody(&s3.GetObjectOutput{
		Body:          io.NopCloser(bytes.NewReader(buf.Bytes())),
		ContentLength: aws.Int64(int64(buf.Len())),
	})
	if err != nil {
		t.Fatalf("parquetRowsFromBody error: %v", err)
	}
	if len(out) != 1 || out[0].ID != 1 {
		t.Fatalf("unexpected parquet rows: %+v", out)
	}
}

func TestS3Client_ParquetRowsFromBodySpill(t *testing.T) {
	var buf bytes.Buffer
	pw := parquet.NewGenericWriter[sampleRow](&buf)
	_, err := pw.Write([]sampleRow{{ID: 2, Name: "b"}})
	if err != nil {
		t.Fatalf("parquet write error: %v", err)
	}
	if err := pw.Close(); err != nil {
		t.Fatalf("parquet close error: %v", err)
	}

	client := NewS3ClientAdapter[sampleRow](context.Background()).(*S3Client[sampleRow])
	client.readerFormatOpts = map[string]string{"spill_threshold_bytes": "1"}
	out, err := client.parquetRowsFromBody(&s3.GetObjectOutput{
		Body:          io.NopCloser(bytes.NewReader(buf.Bytes())),
		ContentLength: aws.Int64(int64(buf.Len())),
	})
	if err != nil {
		t.Fatalf("parquetRowsFromBody error: %v", err)
	}
	if len(out) != 1 || out[0].ID != 2 {
		t.Fatalf("unexpected parquet rows: %+v", out)
	}
}

func TestS3Client_ParquetRoller(t *testing.T) {
	client := NewS3ClientAdapter[sampleRow](context.Background()).(*S3Client[sampleRow])
	in := make(chan sampleRow, 2)
	out := make(chan []byte, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = client.parquetRoller(ctx, in, out, 10*time.Millisecond, 1, parquet.Compression(&parquet.Snappy), "snappy")
	}()

	in <- sampleRow{ID: 1, Name: "a"}
	close(in)

	select {
	case data := <-out:
		if len(data) == 0 {
			t.Fatalf("expected parquet payload")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for parquet output")
	}
	wg.Wait()
}

func TestS3Client_BackoffAndRetryable(t *testing.T) {
	if d := backoffDuration(10); d > defaultMaxBackoff {
		t.Fatalf("expected backoff <= max, got %s", d)
	}
	if !isRetryable(fmt.Errorf("service unavailable")) {
		t.Fatalf("expected retryable error")
	}
	if isRetryable(context.Canceled) {
		t.Fatalf("expected context canceled to be non-retryable")
	}
}

func TestS3Client_FanIn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	src := make(chan int, 1)
	dst := make(chan int, 1)

	client := NewS3ClientAdapter[int](context.Background()).(*S3Client[int])
	go client.fanIn(ctx, dst, src)

	src <- 7
	if got := <-dst; got != 7 {
		t.Fatalf("expected value to be forwarded, got %d", got)
	}
	close(src)
}

func TestS3Client_ServeWriterErrors(t *testing.T) {
	client := NewS3ClientAdapter[int](context.Background()).(*S3Client[int])
	if err := client.ServeWriter(context.Background(), make(chan int)); err == nil {
		t.Fatalf("expected missing deps error")
	}
	if err := client.ServeWriterRaw(context.Background(), make(chan []byte)); err == nil {
		t.Fatalf("expected missing deps error")
	}
	if err := client.Serve(context.Background(), func(context.Context, int) error { return nil }); err == nil {
		t.Fatalf("expected missing deps error")
	}
}

func TestS3Client_SetS3ClientDeps(t *testing.T) {
	client := NewS3ClientAdapter[int](context.Background()).(*S3Client[int])
	client.SetS3ClientDeps(types.S3ClientDeps{
		Client: newTestS3Client(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("unexpected request")
		})),
		Bucket: "  bucket ",
	})
	if client.bucket != "bucket" {
		t.Fatalf("expected bucket to be trimmed, got %q", client.bucket)
	}
}

func TestS3Client_StartWriterValidation(t *testing.T) {
	client := NewS3ClientAdapter[int](context.Background()).(*S3Client[int])
	if err := client.StartWriter(context.Background()); err == nil {
		t.Fatalf("expected missing deps error")
	}

	client.cli = newTestS3Client(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(http.StatusOK, nil, map[string]string{}), nil
	}))
	client.bucket = "bucket"
	if err := client.StartWriter(context.Background()); err == nil {
		t.Fatalf("expected missing wire error")
	}
}

func TestS3Client_RenderKeyDeterministic(t *testing.T) {
	client := NewS3ClientAdapter[int](context.Background()).(*S3Client[int])
	client.prefixTemplate = "prefix"
	client.fileNameTmpl = "file"
	got := client.renderKey(time.Unix(0, 0))
	want := path.Join("prefix", "file")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestS3Client_ReaderConfigCompression(t *testing.T) {
	client := NewS3ClientAdapter[int](context.Background()).(*S3Client[int])
	client.SetReaderConfig(types.S3ReaderConfig{
		Compression: "gzip",
	})
	if client.readerFormatOpts["gzip"] != "true" {
		t.Fatalf("expected gzip option to be set")
	}
}
