package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

// High-volume pipeline item (hot path: keep it allocation-light)
type Feedback struct {
	Seq        uint64
	CustomerID string
	Content    []byte
	Category   string
	IsNegative bool
	Hash       [32]byte
}

// Low-volume telemetry sample (cold path: OK to allocate/log)
type Telemetry struct {
	Seq      uint64
	Cust     string
	Category string
	Hash8    string
	Neg      bool
}

func makePayloadBank(payloadSize, bankItems int) []byte {
	b := make([]byte, payloadSize*bankItems)
	var x uint64 = 0x9E3779B97F4A7C15
	for i := 0; i < len(b); i++ {
		x ^= x >> 12
		x ^= x << 25
		x ^= x >> 27
		b[i] = byte((x * 0x2545F4914F6CDD1D) >> 56)
	}
	return b
}

func main() {
	var (
		items       = flag.Int("items", 10_000_000, "total items")
		workers     = flag.Int("workers", runtime.GOMAXPROCS(0), "worker goroutines (maxConcurrency)")
		queue       = flag.Int("queue", 65_536, "queue depth (bufferSize)")
		payloadSize = flag.Int("payload", 256, "payload bytes")
		bankItems   = flag.Int("bank", 65_536, "payload bank items")
		errorEvery  = flag.Int("errorEvery", 0, "drop as error every N (0 disables)")
		rounds      = flag.Int("rounds", 1, "sha rounds (>=1)")
		sampleEvery = flag.Int("sampleEvery", 8192, "telemetry sampling every N (0 disables)")
		timeout     = flag.Duration("timeout", 60*time.Second, "hard timeout")
	)
	flag.Parse()

	if *items <= 0 || *workers <= 0 || *queue <= 0 || *payloadSize <= 0 || *bankItems <= 0 || *rounds <= 0 {
		panic("bad flags")
	}

	fmt.Printf("ENV: go=%s CPUs=%d GOMAXPROCS=%d\n", runtime.Version(), runtime.NumCPU(), runtime.GOMAXPROCS(0))
	fmt.Printf("CONFIG: items=%d workers=%d queue=%d payload=%d bank=%d rounds=%d errorEvery=%d sampleEvery=%d\n",
		*items, *workers, *queue, *payloadSize, *bankItems, *rounds, *errorEvery, *sampleEvery,
	)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Logger (used on telemetry path only; hot path stays clean)
	_ = os.MkdirAll("logs", 0o755)
	logger := builder.NewLogger(
		builder.LoggerWithDevelopment(true),
		builder.LoggerWithLevel("info"),
	)
	defer logger.Flush()

	_ = logger.AddSink("stdout", builder.SinkConfig{Type: "stdout"})
	_ = logger.AddSink("file", builder.SinkConfig{
		Type: string(builder.FileSink),
		Config: map[string]interface{}{
			"path": "logs/feedback_pipeline.log",
		},
	})

	// Payload bank for hot path (no per-item content alloc)
	payloadBank := makePayloadBank(*payloadSize, *bankItems)

	// Hashing: stack fast path for payload=256, pooled fallback otherwise.
	bufLen := *payloadSize + 8 + 32
	bufPool := sync.Pool{New: func() any { return make([]byte, bufLen) }}

	hashFeedback := func(seq uint64, payload []byte) [32]byte {
		// Fast stack buffer when payload=256 (best speed).
		if *payloadSize == 256 {
			var buf [256 + 8 + 32]byte
			copy(buf[:256], payload)
			binary.LittleEndian.PutUint64(buf[256:264], seq)

			var prev [32]byte
			for i := 0; i < *rounds; i++ {
				copy(buf[264:], prev[:])
				prev = sha256.Sum256(buf[:])
			}
			return prev
		}

		// Generic pooled buffer for other payload sizes.
		b := bufPool.Get().([]byte)
		if cap(b) < bufLen {
			b = make([]byte, bufLen)
		}
		b = b[:bufLen]
		defer bufPool.Put(b)

		copy(b[:*payloadSize], payload)
		binary.LittleEndian.PutUint64(b[*payloadSize:*payloadSize+8], seq)

		var prev [32]byte
		for i := 0; i < *rounds; i++ {
			copy(b[*payloadSize+8:], prev[:])
			prev = sha256.Sum256(b)
		}
		return prev
	}

	// Hot-path transform: one function so Wire can take the fast loop.
	var errCount uint64
	transformFeedback := func(f Feedback) (Feedback, error) {
		// deterministic error injection
		if *errorEvery > 0 && (f.Seq%uint64(*errorEvery) == 0) {
			atomic.AddUint64(&errCount, 1)
			return Feedback{}, fmt.Errorf("simulated error")
		}

		// classify (deterministic, no map iteration)
		// NOTE: payloadBank is already pseudo-random bytes; for a real demo you’d use real text.
		// We'll simulate “negative” and “category” based on a few byte checks.
		if len(f.Content) > 0 && (f.Content[0]&1) == 1 {
			f.IsNegative = true
			f.Category = "Support"
		} else {
			f.IsNegative = false
			f.Category = "Delivery"
		}

		f.Hash = hashFeedback(f.Seq, f.Content)
		return f, nil
	}

	// Telemetry wire: sampled, instrumented, logged (cold path).
	telemetryCh := make(chan Telemetry, 65_536)

	var telemetryCount uint64
	telemetrySensor := builder.NewSensor[Telemetry](
		builder.SensorWithOnElementProcessedFunc[Telemetry](func(_ builder.ComponentMetadata, _ Telemetry) {
			atomic.AddUint64(&telemetryCount, 1)
		}),
	)

	telemetryPlug := builder.NewPlug[Telemetry](
		ctx,
		builder.PlugWithAdapterFunc[Telemetry](func(ctx context.Context, submit func(context.Context, Telemetry) error) {
			for {
				select {
				case <-ctx.Done():
					return
				case t := <-telemetryCh:
					_ = submit(ctx, t)
				}
			}
		}),
	)

	telemetryGen := builder.NewGenerator[Telemetry](
		ctx,
		builder.GeneratorWithPlug[Telemetry](telemetryPlug),
	)

	telemetryTransform := func(t Telemetry) (Telemetry, error) {
		// This is intentionally “expensive-ish” because it’s sampled.
		logger.Info("telemetry",
			"seq", t.Seq,
			"cust", t.Cust,
			"cat", t.Category,
			"neg", t.Neg,
			"hash8", t.Hash8,
		)
		return t, nil
	}

	telemetryWire := builder.NewWire[Telemetry](
		ctx,
		builder.WireWithTransformer[Telemetry](telemetryTransform),
		builder.WireWithGenerator[Telemetry](telemetryGen),
		builder.WireWithSensor[Telemetry](telemetrySensor),
		builder.WireWithConcurrencyControl[Telemetry](4096, 2),
	)

	// Drain telemetry output (we don't need it).
	go func() {
		for range telemetryWire.GetOutputChannel() {
		}
	}()

	_ = telemetryWire.Start(ctx)

	// Processing wire: max throughput (no logger/sensor attached).
	feedbackPlug := builder.NewPlug[Feedback](
		ctx,
		builder.PlugWithAdapterFunc[Feedback](func(ctx context.Context, submit func(context.Context, Feedback) error) {
			for i := 0; i < *items; i++ {
				start := (i % *bankItems) * (*payloadSize)
				f := Feedback{
					Seq:        uint64(i),
					CustomerID: "C" + strconv.Itoa(i%10000),
					Content:    payloadBank[start : start+(*payloadSize)],
				}
				if err := submit(ctx, f); err != nil || ctx.Err() != nil {
					return
				}
			}
		}),
	)

	feedbackGen := builder.NewGenerator[Feedback](
		ctx,
		builder.GeneratorWithPlug[Feedback](feedbackPlug),
	)

	processingWire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithTransformer[Feedback](transformFeedback),
		builder.WireWithGenerator[Feedback](feedbackGen),
		builder.WireWithConcurrencyControl[Feedback](*queue, *workers),
	)

	// Completion: ok + err == items
	var okCount uint64
	done := make(chan struct{})
	var doneOnce sync.Once
	maybeDone := func() {
		if atomic.LoadUint64(&okCount)+atomic.LoadUint64(&errCount) == uint64(*items) {
			doneOnce.Do(func() { close(done) })
		}
	}

	// Drain output, count, and sample telemetry without slowing hot path.
	go func() {
		for f := range processingWire.GetOutputChannel() {
			atomic.AddUint64(&okCount, 1)

			// Sampling: try-send so telemetry never back-pressures the hot path.
			if *sampleEvery > 0 && (f.Seq%uint64(*sampleEvery) == 0) {
				// Create a tiny hash prefix string for log readability (cold path).
				h := hex.EncodeToString(f.Hash[:4])
				t := Telemetry{
					Seq:      f.Seq,
					Cust:     f.CustomerID,
					Category: f.Category,
					Hash8:    h,
					Neg:      f.IsNegative,
				}
				select {
				case telemetryCh <- t:
				default:
				}
			}

			maybeDone()
		}
		doneOnce.Do(func() { close(done) })
	}()

	start := time.Now()
	_ = processingWire.Start(ctx)

	select {
	case <-done:
	case <-ctx.Done():
	}

	_ = processingWire.Stop()
	_ = telemetryWire.Stop()
	elapsed := time.Since(start)

	ok := atomic.LoadUint64(&okCount)
	er := atomic.LoadUint64(&errCount)
	rate := float64(ok+er) / elapsed.Seconds()

	logger.Info("summary",
		"ok", ok,
		"err", er,
		"telemetry", atomic.LoadUint64(&telemetryCount),
		"elapsed", elapsed.Truncate(time.Millisecond).String(),
		"rate_items_s", fmt.Sprintf("%.0f", rate),
		"note", strings.Join([]string{
			"hot path: no per-item logging/sensor callbacks",
			"telemetry: sampled + logged via separate wire",
		}, "; "),
	)

	fmt.Printf("RESULT: ok=%d err=%d telemetry=%d elapsed=%s rate=%.0f items/s\n",
		ok, er, atomic.LoadUint64(&telemetryCount), elapsed.Truncate(time.Millisecond), rate,
	)
}
