package kafkaclient

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/segmentio/kafka-go"
)

type testRecord struct {
	CustomerID string `json:"customerId"`
	Count      int    `json:"count"`
}

func TestRenderKeyFromTemplateField(t *testing.T) {
	rec := testRecord{CustomerID: "C-01", Count: 3}

	key := renderKeyFromTemplate("{customerId}", rec)
	if string(key) != "C-01" {
		t.Fatalf("expected key to render from field, got %q", string(key))
	}
}

func TestRenderKeyFromTemplateLiteral(t *testing.T) {
	key := renderKeyFromTemplate("static-key", testRecord{})
	if string(key) != "static-key" {
		t.Fatalf("expected literal key, got %q", string(key))
	}
}

func TestRenderHeadersFromTemplates(t *testing.T) {
	rec := testRecord{CustomerID: "C-42"}
	headers := renderHeadersFromTemplates(map[string]string{
		"source":   "demo",
		"customer": "{customerId}",
	}, rec)

	if len(headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(headers))
	}

	vals := make(map[string]string, len(headers))
	for _, h := range headers {
		vals[h.Key] = h.Value
	}

	if vals["source"] != "demo" {
		t.Fatalf("expected source header, got %q", vals["source"])
	}
	if vals["customer"] != "C-42" {
		t.Fatalf("expected customer header, got %q", vals["customer"])
	}
}

func TestRenderFieldFromValueRejectsInvalidPlaceholder(t *testing.T) {
	if _, ok := renderFieldFromValue(testRecord{}, "customerId"); ok {
		t.Fatalf("expected invalid placeholder to return false")
	}
}

func TestSetComponentMetadataPreservesType(t *testing.T) {
	adapter := NewKafkaClientAdapter[int](context.Background()).(*KafkaClient[int])
	adapter.SetComponentMetadata("demo", "id-1")
	if adapter.componentMetadata.Type != adapter.Name() {
		t.Fatalf("expected Type %q, got %q", adapter.Name(), adapter.componentMetadata.Type)
	}
}

func TestSetWriterConfig(t *testing.T) {
	adapter := NewKafkaClientAdapter[int](context.Background()).(*KafkaClient[int])
	p := 2
	cfg := types.KafkaWriterConfig{
		Topic:             "topic-a",
		Format:            "ndjson",
		FormatOptions:     map[string]string{"foo": "bar"},
		Compression:       "gzip",
		KeyTemplate:       "{customerId}",
		HeaderTemplates:   map[string]string{"source": "demo"},
		BatchMaxRecords:   10,
		BatchMaxBytes:     20,
		BatchMaxAge:       2 * time.Second,
		Acks:              "all",
		RequestTimeout:    3 * time.Second,
		PartitionStrategy: "hash",
		ManualPartition:   &p,
		EnableDLQ:         true,
	}
	adapter.SetWriterConfig(cfg)

	if adapter.wTopic != "topic-a" {
		t.Fatalf("expected topic to be set")
	}
	if adapter.wFormatOpts["foo"] != "bar" {
		t.Fatalf("expected format option to be merged")
	}
	if adapter.wHdrTemplates["source"] != "demo" {
		t.Fatalf("expected header template to be merged")
	}
	if adapter.wBatchMaxRecords != 10 {
		t.Fatalf("expected batch max records to be set")
	}
	if adapter.wManualPartition == nil || *adapter.wManualPartition != 2 {
		t.Fatalf("expected manual partition to be set")
	}
}

func TestSetReaderConfig(t *testing.T) {
	adapter := NewKafkaClientAdapter[int](context.Background()).(*KafkaClient[int])
	ts := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	cfg := types.KafkaReaderConfig{
		GroupID:        "group-a",
		Topics:         []string{"topic-a"},
		StartAt:        "timestamp",
		StartAtTime:    ts,
		PollInterval:   500 * time.Millisecond,
		MaxPollRecords: 12,
		MaxPollBytes:   1 << 12,
		Format:         "ndjson",
		FormatOptions:  map[string]string{"a": "b"},
		CommitMode:     "manual",
		CommitPolicy:   "time",
		CommitInterval: 2 * time.Second,
	}
	adapter.SetReaderConfig(cfg)

	if adapter.rGroupID != "group-a" {
		t.Fatalf("expected group to be set")
	}
	if len(adapter.rTopics) != 1 || adapter.rTopics[0] != "topic-a" {
		t.Fatalf("expected topics to be set")
	}
	if !adapter.rStartAtTime.Equal(ts) {
		t.Fatalf("expected start time to be set")
	}
	if adapter.rCommitMode != "manual" {
		t.Fatalf("expected commit mode to be set")
	}
}

func TestEffectiveWriterTopic(t *testing.T) {
	adapter := NewKafkaClientAdapter[int](context.Background()).(*KafkaClient[int])
	adapter.wTopic = "topic-a"
	if topic, ok := adapter.effectiveWriterTopic(); !ok || topic != "topic-a" {
		t.Fatalf("expected configured topic to win")
	}

	adapter.wTopic = ""
	adapter.producer = &kafka.Writer{Topic: "topic-b"}
	if topic, ok := adapter.effectiveWriterTopic(); !ok || topic != "topic-b" {
		t.Fatalf("expected writer topic to be used")
	}

	adapter.producer = nil
	if _, ok := adapter.effectiveWriterTopic(); ok {
		t.Fatalf("expected no topic when unset")
	}
}

func TestServeWriterRequiresTopic(t *testing.T) {
	adapter := NewKafkaClientAdapter[int](context.Background()).(*KafkaClient[int])
	ch := make(chan int)
	close(ch)

	if err := adapter.ServeWriter(context.Background(), ch); err == nil {
		t.Fatalf("expected error when no topic is configured")
	}
}

func TestServeWriterRejectsUnsupportedFormat(t *testing.T) {
	adapter := NewKafkaClientAdapter[int](context.Background()).(*KafkaClient[int])
	adapter.SetWriterConfig(types.KafkaWriterConfig{Topic: "topic-a", Format: "xml"})
	ch := make(chan int)
	close(ch)

	err := adapter.ServeWriter(context.Background(), ch)
	if err == nil || !strings.Contains(err.Error(), "ndjson") {
		t.Fatalf("expected ndjson format error, got %v", err)
	}
}

func TestServeWriterRawRequiresTopic(t *testing.T) {
	adapter := NewKafkaClientAdapter[int](context.Background()).(*KafkaClient[int])
	ch := make(chan []byte)
	close(ch)

	if err := adapter.ServeWriterRaw(context.Background(), ch); err == nil {
		t.Fatalf("expected error when no topic is configured")
	}
}
