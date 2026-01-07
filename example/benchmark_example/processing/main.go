package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"hash"
	"runtime"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Event struct {
	Seq        uint64
	Payload    []byte
	IsNegative bool
	Hash       [32]byte
}

var errSimulated = errors.New("simulated error")

// Deterministic payload bank (no per-item allocations).
func makePayloadBank(payloadSize, bankItems int) []byte {
	b := make([]byte, payloadSize*bankItems)
	// xorshift64*
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
		workers     = flag.Int("workers", 32, "wire maxConcurrency (worker pool size)")
		queue       = flag.Int("queue", 65_536, "wire buffer size (queue depth)")
		payloadSize = flag.Int("payload", 256, "payload size in bytes")
		bankItems   = flag.Int("bank", 65_536, "number of payload slices in bank")
		rounds      = flag.Int("rounds", 1, "hash rounds (each round hashes payload + seq + prev hash)")
		errorEvery  = flag.Int("errorEvery", 0, "inject error on every Nth item (0 disables)")
	)
	flag.Parse()

	if *payloadSize <= 0 {
		panic("payload must be > 0")
	}
	if *bankItems <= 0 {
		panic("bank must be > 0")
	}
	if *rounds <= 0 {
		panic("rounds must be > 0")
	}

	fmt.Printf("CONFIG: items=%d workers=%d queue=%d payload=%dB bank=%d rounds=%d errorEvery=%d GOMAXPROCS=%d CPUs=%d\n",
		*items, *workers, *queue, *payloadSize, *bankItems, *rounds, *errorEvery, runtime.GOMAXPROCS(0), runtime.NumCPU(),
	)

	payloadBank := makePayloadBank(*payloadSize, *bankItems)

	// Transformers ------------------------------------------------------------

	injectError := func(e Event) (Event, error) {
		if *errorEvery > 0 && (e.Seq%uint64(*errorEvery) == 0) {
			return Event{}, errSimulated
		}
		return e, nil
	}

	// Pool sha256 hashers to avoid per-item allocations.
	hasherPool := sync.Pool{
		New: func() any { return sha256.New() },
	}

	hashProcessor := func(e Event) (Event, error) {
		h := hasherPool.Get().(hash.Hash)
		defer hasherPool.Put(h)

		var sum [32]byte
		var tmp [32]byte
		var seqBuf [8]byte
		binary.LittleEndian.PutUint64(seqBuf[:], e.Seq)

		for i := 0; i < *rounds; i++ {
			h.Reset()
			_, _ = h.Write(e.Payload)
			_, _ = h.Write(seqBuf[:])
			_, _ = h.Write(sum[:])
			out := h.Sum(tmp[:0]) // no alloc; writes into tmp
			copy(sum[:], out)
		}

		e.Hash = sum
		return e, nil
	}

	// Plug / generator --------------------------------------------------------

	plug := func(ctx context.Context, submit func(context.Context, Event) error) {
		for i := 0; i < *items; i++ {
			start := (i % *bankItems) * (*payloadSize)
			ev := Event{
				Seq:        uint64(i),
				Payload:    payloadBank[start : start+(*payloadSize)],
				IsNegative: (i%10 == 0),
			}
			if err := submit(ctx, ev); err != nil || ctx.Err() != nil {
				return
			}
		}
	}

	// Wire --------------------------------------------------------------------

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	meter := builder.NewMeter[Event](ctx,
		builder.MeterWithTotalItems[Event](uint64(*items)),
	)

	wire := builder.NewWire(
		ctx,
		// Error first (cheap fail) then CPU work
		builder.WireWithTransformer(injectError, hashProcessor),
		builder.WireWithGenerator(
			builder.NewGenerator(ctx,
				builder.GeneratorWithPlug(
					builder.NewPlug(ctx, builder.PlugWithAdapterFunc(plug)),
				),
			),
		),
		builder.WireWithSensor(builder.NewSensor(builder.SensorWithMeter[Event](meter))),
		builder.WireWithConcurrencyControl[Event](*queue, *workers),
	)

	// Drain output so queue depth can be realistic (small) without stalling.
	go func() {
		for range wire.GetOutputChannel() {
		}
	}()

	startTime := time.Now()
	_ = wire.Start(ctx)
	meter.Monitor()
	_ = wire.Stop()

	elapsed := time.Since(startTime)

	fmt.Printf("\nRESULT: processed=%d elapsed=%s rate=%.0f items/s\n",
		*items, elapsed.Truncate(time.Millisecond), float64(*items)/elapsed.Seconds())
}
