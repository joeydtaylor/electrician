//go:build integration
// +build integration

package kafkaclient_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
	"github.com/segmentio/kafka-go"
)

type kafkaIntegrationMsg struct {
	ID  string `json:"id"`
	Seq int    `json:"seq"`
}

func TestKafkaClientLocalstackRoundTrip(t *testing.T) {
	if os.Getenv("SKIP_KAFKA") == "1" {
		t.Skip("SKIP_KAFKA=1")
	}

	brokers := splitCSV(envOr("KAFKA_BROKERS", "127.0.0.1:19092"))
	if len(brokers) == 0 {
		t.Fatal("KAFKA_BROKERS is empty")
	}

	root := repoRoot(t)
	caCSV := resolvePath(root, envOr("KAFKA_TLS_CA", filepath.Join("local-stack", "tls", "ca.crt")))
	certCSV := resolvePath(root, envOr("KAFKA_TLS_CERT", filepath.Join("local-stack", "tls", "client.crt")))
	keyCSV := resolvePath(root, envOr("KAFKA_TLS_KEY", filepath.Join("local-stack", "tls", "client.key")))
	serverName := envOr("KAFKA_TLS_SERVER_NAME", "localhost")
	tlsCfg, err := builder.TLSFromMTLSPathCSV(caCSV, certCSV, keyCSV, serverName)
	if err != nil {
		t.Fatalf("load kafka tls config: %v", err)
	}

	user := envOr("KAFKA_SASL_USER", "app")
	pass := envOr("KAFKA_SASL_PASS", "app-secret")
	mechName := envOr("KAFKA_SASL_MECH", "SCRAM-SHA-256")
	mech, err := builder.SASLSCRAM(user, pass, mechName)
	if err != nil {
		t.Fatalf("build sasl: %v", err)
	}

	dialer := builder.NewKafkaGoDialer(tlsCfg, mech, 5*time.Second, true)
	if err := probeKafka(dialer, brokers[0]); err != nil {
		t.Fatalf("kafka broker not reachable (%s): %v", brokers[0], err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	topic := envOr("KAFKA_TOPIC", "feedback-demo")
	groupID := fmt.Sprintf("electrician-it-%d", time.Now().UnixNano())

	writer := builder.NewKafkaGoWriterSecure(brokers, topic, tlsCfg, mech, "electrician-it")
	defer func() { _ = writer.Close() }()

	reader := builder.NewKafkaGoReaderSecure(
		brokers,
		groupID,
		[]string{topic},
		tlsCfg,
		mech,
		10*time.Second,
		true,
		builder.KafkaGoReaderWithEarliestStart(),
	)
	defer func() { _ = reader.Close() }()

	adapter := builder.NewKafkaClientAdapter[kafkaIntegrationMsg](
		ctx,
		builder.KafkaClientAdapterWithKafkaGoWriter[kafkaIntegrationMsg](writer),
		builder.KafkaClientAdapterWithKafkaGoReader[kafkaIntegrationMsg](reader),
		builder.KafkaClientAdapterWithWriterTopic[kafkaIntegrationMsg](topic),
		builder.KafkaClientAdapterWithReaderTopics[kafkaIntegrationMsg](topic),
		builder.KafkaClientAdapterWithReaderGroup[kafkaIntegrationMsg](groupID),
		builder.KafkaClientAdapterWithReaderStartAt[kafkaIntegrationMsg]("earliest", time.Time{}),
		builder.KafkaClientAdapterWithReaderPollSettings[kafkaIntegrationMsg](250*time.Millisecond, 1000, 1<<20),
	)

	testID := fmt.Sprintf("it-%d", time.Now().UnixNano())
	want := []kafkaIntegrationMsg{
		{ID: testID, Seq: 1},
		{ID: testID, Seq: 2},
		{ID: testID, Seq: 3},
	}

	gotCh := make(chan kafkaIntegrationMsg, len(want))
	serveErr := make(chan error, 1)
	serveCtx, serveCancel := context.WithTimeout(ctx, 30*time.Second)
	go func() {
		serveErr <- adapter.Serve(serveCtx, func(_ context.Context, v kafkaIntegrationMsg) error {
			if v.ID != testID {
				return nil
			}
			select {
			case gotCh <- v:
			default:
			}
			if len(gotCh) >= len(want) {
				serveCancel()
			}
			return nil
		})
	}()

	in := make(chan kafkaIntegrationMsg, len(want))
	writeErr := make(chan error, 1)
	go func() { writeErr <- adapter.ServeWriter(ctx, in) }()
	for _, m := range want {
		in <- m
	}
	close(in)

	if err := <-writeErr; err != nil {
		serveCancel()
		t.Fatalf("write failed: %v", err)
	}

	got := make([]kafkaIntegrationMsg, 0, len(want))
	timeout := time.After(30 * time.Second)
	for len(got) < len(want) {
		select {
		case v := <-gotCh:
			got = append(got, v)
		case err := <-serveErr:
			if err != nil && err != context.Canceled {
				t.Fatalf("serve failed: %v", err)
			}
		case <-timeout:
			serveCancel()
			t.Fatalf("timed out waiting for kafka messages (got %d/%d)", len(got), len(want))
		}
	}

	select {
	case err := <-serveErr:
		if err != nil && err != context.Canceled {
			t.Fatalf("serve failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		serveCancel()
	}

	wantSeq := map[int]bool{}
	for _, m := range want {
		wantSeq[m.Seq] = true
	}
	for _, m := range got {
		if m.ID != testID {
			continue
		}
		if !wantSeq[m.Seq] {
			t.Fatalf("unexpected message: %+v", m)
		}
		wantSeq[m.Seq] = false
	}
	for seq, ok := range wantSeq {
		if ok {
			t.Fatalf("missing message seq=%d", seq)
		}
	}
}

func probeKafka(dialer *kafka.Dialer, broker string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := dialer.DialContext(ctx, "tcp", broker)
	if err != nil {
		return err
	}
	return conn.Close()
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("repo root not found (go.mod)")
		}
		dir = parent
	}
}

func resolvePath(root, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(root, p)
}

func splitCSV(csv string) []string {
	var out []string
	for _, part := range strings.Split(csv, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
