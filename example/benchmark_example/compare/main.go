package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"
	"sync/atomic"
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
	var x uint64 = 0x9E3779B97F4A7C15
	for i := 0; i < len(b); i++ {
		x ^= x >> 12
		x ^= x << 25
		x ^= x >> 27
		b[i] = byte((x * 0x2545F4914F6CDD1D) >> 56)
	}
	return b
}

type Sink interface {
	Write(e Event) error
	Close() error
}

type DiscardSink struct{}

func (d DiscardSink) Write(Event) error { return nil }
func (d DiscardSink) Close() error      { return nil }

type FileSink struct {
	f  *os.File
	bw *bufio.Writer
}

func NewFileSink(path string) (*FileSink, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	bw := bufio.NewWriterSize(f, 1<<20) // 1MB buffer
	return &FileSink{f: f, bw: bw}, nil
}

func (s *FileSink) Write(e Event) error {
	// Zero-alloc line formatting:
	// "<seq> <hashhex>\n"
	var buf [128]byte
	b := buf[:0]

	b = strconv.AppendUint(b, e.Seq, 10)
	b = append(b, ' ')

	start := len(b)
	b = b[:start+64]
	hex.Encode(b[start:], e.Hash[:])

	b = append(b, '\n')

	_, err := s.bw.Write(b)
	return err
}

func (s *FileSink) Close() error {
	if s == nil {
		return nil
	}
	if err := s.bw.Flush(); err != nil {
		_ = s.f.Close()
		return err
	}
	return s.f.Close()
}

func main() {
	var (
		engine      = flag.String("engine", "electrician", "electrician|baseline")
		sinkMode    = flag.String("sink", "discard", "discard|file")
		outPath     = flag.String("out", "out.log", "file path when -sink=file")
		items       = flag.Int("items", 10_000_000, "total items")
		workers     = flag.Int("workers", runtime.GOMAXPROCS(0), "worker goroutines")
		queue       = flag.Int("queue", 65_536, "queue depth")
		payloadSize = flag.Int("payload", 256, "payload bytes")
		bankItems   = flag.Int("bank", 65_536, "payload bank items")
		rounds      = flag.Int("rounds", 1, "sha rounds")
		errorEvery  = flag.Int("errorEvery", 0, "inject error every N (0 disables)")
		timeout     = flag.Duration("timeout", 20*time.Minute, "hard timeout")
		cpuprofile  = flag.String("cpuprofile", "", "write CPU profile to file (e.g. elec.cpu.pprof)")
	)
	flag.Parse()

	if *items <= 0 || *workers <= 0 || *queue <= 0 || *payloadSize <= 0 || *bankItems <= 0 || *rounds <= 0 {
		panic("bad flags")
	}

	fmt.Printf("ENV: go=%s CPUs=%d GOMAXPROCS=%d\n", runtime.Version(), runtime.NumCPU(), runtime.GOMAXPROCS(0))
	fmt.Printf("CONFIG: engine=%s sink=%s items=%d workers=%d queue=%d payload=%d bank=%d rounds=%d errorEvery=%d\n",
		*engine, *sinkMode, *items, *workers, *queue, *payloadSize, *bankItems, *rounds, *errorEvery,
	)

	payloadBank := makePayloadBank(*payloadSize, *bankItems)

	// Sink
	var sink Sink = DiscardSink{}
	if *sinkMode == "file" {
		fs, err := NewFileSink(*outPath)
		if err != nil {
			panic(err)
		}
		sink = fs
	}
	defer func() { _ = sink.Close() }()

	// Completion accounting
	var okCount uint64
	var errCount uint64
	done := make(chan struct{})
	var doneOnce sync.Once

	maybeDone := func() {
		if atomic.LoadUint64(&okCount)+atomic.LoadUint64(&errCount) == uint64(*items) {
			doneOnce.Do(func() { close(done) })
		}
	}

	// Error injection (optional)
	injectError := func(e Event) (Event, error) {
		if *errorEvery > 0 && (e.Seq%uint64(*errorEvery) == 0) {
			atomic.AddUint64(&errCount, 1)
			maybeDone()
			return Event{}, errSimulated
		}
		return e, nil
	}

	// runtime stats
	var peakGo int
	var peakHeap uint64
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	startAlloc := ms.TotalAlloc
	startGC := ms.NumGC

	monStop := make(chan struct{})
	go func() {
		t := time.NewTicker(200 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-monStop:
				return
			case <-t.C:
				g := runtime.NumGoroutine()
				if g > peakGo {
					peakGo = g
				}
				runtime.ReadMemStats(&ms)
				if ms.HeapInuse > peakHeap {
					peakHeap = ms.HeapInuse
				}
			}
		}
	}()

	// CPU profiling (optional)
	var cpuFile *os.File
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		cpuFile = f
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			_ = cpuFile.Close()
			panic(err)
		}
	}

	start := time.Now()

	switch *engine {
	case "electrician":
		runElectrician(
			*items, *workers, *queue, *payloadSize, *bankItems, *rounds,
			payloadBank, *errorEvery, *timeout,
			injectError, sink,
			&okCount, &errCount, maybeDone, done, &doneOnce,
		)
	case "baseline":
		runBaseline(
			*items, *workers, *queue, *payloadSize, *bankItems, *rounds,
			payloadBank, *errorEvery, *timeout,
			injectError, sink,
			&okCount, &errCount, maybeDone, done, &doneOnce,
		)
	default:
		panic("engine must be electrician or baseline")
	}

	elapsed := time.Since(start)

	if cpuFile != nil {
		pprof.StopCPUProfile()
		_ = cpuFile.Close()
	}

	close(monStop)

	runtime.ReadMemStats(&ms)
	allocMB := float64(ms.TotalAlloc-startAlloc) / (1024 * 1024)
	numGC := ms.NumGC - startGC

	ok := atomic.LoadUint64(&okCount)
	er := atomic.LoadUint64(&errCount)
	total := ok + er

	fmt.Printf(
		"RESULT: ok=%d err=%d total=%d elapsed=%s rate=%.0f items/s peak_go=%d peak_heap_mb=%.1f alloc_mb=%.1f gc=%d\n",
		ok, er, total, elapsed.Truncate(time.Millisecond),
		float64(total)/elapsed.Seconds(),
		peakGo, float64(peakHeap)/(1024*1024), allocMB, numGC,
	)
}

func runElectrician(
	items, workers, queue, payloadSize, bankItems, rounds int,
	payloadBank []byte,
	errorEvery int,
	timeout time.Duration,
	injectError func(Event) (Event, error),
	sink Sink,
	okCount, errCount *uint64,
	maybeDone func(),
	done chan struct{},
	doneOnce *sync.Once,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	plug := func(ctx context.Context, submit func(context.Context, Event) error) {
		for i := 0; i < items; i++ {
			start := (i % bankItems) * payloadSize
			ev := Event{
				Seq:        uint64(i),
				Payload:    payloadBank[start : start+payloadSize],
				IsNegative: (i%10 == 0),
			}
			if err := submit(ctx, ev); err != nil || ctx.Err() != nil {
				return
			}
		}
	}

	gen := builder.NewGenerator[Event](
		ctx,
		builder.GeneratorWithPlug[Event](
			builder.NewPlug[Event](ctx, builder.PlugWithAdapterFunc[Event](plug)),
		),
	)

	// Use the new toy: per-worker scratch buffer (allocated once per worker).
	bufLen := payloadSize + 8 + 32

	hashTransformerWithScratch := builder.WireWithScratchBytes[Event](bufLen, func(buf []byte, e Event) (Event, error) {
		// Layout: [payload][seq(8)][prev(32)]
		copy(buf[:payloadSize], e.Payload)
		binary.LittleEndian.PutUint64(buf[payloadSize:payloadSize+8], e.Seq)

		var prev [32]byte
		for i := 0; i < rounds; i++ {
			copy(buf[payloadSize+8:], prev[:])
			prev = sha256.Sum256(buf)
		}
		e.Hash = prev
		return e, nil
	})

	var wire interface {
		Start(context.Context) error
		Stop() error
		GetOutputChannel() chan Event
	}

	if errorEvery > 0 {
		wire = builder.NewWire[Event](
			ctx,
			builder.WireWithTransformer[Event](injectError),
			hashTransformerWithScratch,
			builder.WireWithGenerator[Event](gen),
			builder.WireWithConcurrencyControl[Event](queue, workers),
		)
	} else {
		wire = builder.NewWire[Event](
			ctx,
			hashTransformerWithScratch,
			builder.WireWithGenerator[Event](gen),
			builder.WireWithConcurrencyControl[Event](queue, workers),
		)
	}

	go func() {
		for e := range wire.GetOutputChannel() {
			_ = sink.Write(e)
			atomic.AddUint64(okCount, 1)
			maybeDone()
		}
		doneOnce.Do(func() { close(done) })
	}()

	_ = wire.Start(ctx)

	select {
	case <-done:
	case <-time.After(timeout):
	}

	cancel()
	_ = wire.Stop()
}

func runBaseline(
	items, workers, queue, payloadSize, bankItems, rounds int,
	payloadBank []byte,
	errorEvery int,
	timeout time.Duration,
	injectError func(Event) (Event, error),
	sink Sink,
	okCount, errCount *uint64,
	maybeDone func(),
	done chan struct{},
	doneOnce *sync.Once,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan Event, queue)
	out := make(chan Event, queue)

	// Baseline keeps a per-worker scratch buffer too (fair comparison).
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		scratch := make([]byte, payloadSize+8+32)

		go func(buf []byte) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case e, ok := <-in:
					if !ok {
						return
					}

					if errorEvery > 0 {
						var err error
						e, err = injectError(e)
						if err != nil {
							continue
						}
					}

					// Hash in worker-local scratch (no alloc)
					copy(buf[:payloadSize], e.Payload)
					binary.LittleEndian.PutUint64(buf[payloadSize:payloadSize+8], e.Seq)

					var prev [32]byte
					for r := 0; r < rounds; r++ {
						copy(buf[payloadSize+8:], prev[:])
						prev = sha256.Sum256(buf)
					}
					e.Hash = prev

					select {
					case out <- e:
					case <-ctx.Done():
						return
					}
				}
			}
		}(scratch)
	}

	go func() {
		for e := range out {
			_ = sink.Write(e)
			atomic.AddUint64(okCount, 1)
			maybeDone()
		}
		doneOnce.Do(func() { close(done) })
	}()

	go func() {
		defer close(in)
		for i := 0; i < items; i++ {
			start := (i % bankItems) * payloadSize
			ev := Event{
				Seq:        uint64(i),
				Payload:    payloadBank[start : start+payloadSize],
				IsNegative: (i%10 == 0),
			}
			select {
			case in <- ev:
			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(timeout):
	}

	cancel()
	wg.Wait()
	close(out)
}
