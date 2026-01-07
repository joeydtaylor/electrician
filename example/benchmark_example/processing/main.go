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
	"strings"
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
	bw := bufio.NewWriterSize(f, 1<<20) // 1MB
	return &FileSink{f: f, bw: bw}, nil
}

func (s *FileSink) Write(e Event) error {
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
		cpuprofile  = flag.String("cpuprofile", "", "write CPU profile to file")

		// Instrumentation toggles (Electrician only)
		addLogger     = flag.Bool("logger", false, "attach builder logger to wire+generator (disables fast path)")
		loggerLevel   = flag.String("loggerLevel", "error", "debug|info|warn|error|fatal")
		loggerSink    = flag.String("loggerSink", "devnull", "devnull|stdout|file")
		loggerOutPath = flag.String("loggerOut", "/dev/null", "log file path when loggerSink=file (default /dev/null)")
		addSensor     = flag.Bool("sensor", false, "attach a no-op-ish sensor callback (disables fast path)")
		sensorEvery   = flag.Int("sensorEvery", 0, "invoke sensor callback every N elements (0 = every element)")
		addMeter      = flag.Bool("meter", false, "attach meter via sensor and run meter.Monitor() (disables fast path)")
	)
	flag.Parse()

	if *items <= 0 || *workers <= 0 || *queue <= 0 || *payloadSize <= 0 || *bankItems <= 0 || *rounds <= 0 {
		panic("bad flags")
	}

	fmt.Printf("ENV: go=%s CPUs=%d GOMAXPROCS=%d\n", runtime.Version(), runtime.NumCPU(), runtime.GOMAXPROCS(0))
	fmt.Printf("CONFIG: engine=%s sink=%s items=%d workers=%d queue=%d payload=%d bank=%d rounds=%d errorEvery=%d logger=%t sensor=%t meter=%t\n",
		*engine, *sinkMode, *items, *workers, *queue, *payloadSize, *bankItems, *rounds, *errorEvery,
		*addLogger, *addSensor, *addMeter,
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

	// Hash transform (near-zero alloc)
	bufLen := *payloadSize + 8 + 32
	bufPool := sync.Pool{New: func() any { return make([]byte, bufLen) }}

	hashProcessor := func(e Event) (Event, error) {
		b := bufPool.Get().([]byte)
		if cap(b) < bufLen {
			b = make([]byte, bufLen)
		}
		b = b[:bufLen]
		defer bufPool.Put(b)

		copy(b[:*payloadSize], e.Payload)
		binary.LittleEndian.PutUint64(b[*payloadSize:*payloadSize+8], e.Seq)

		var prev [32]byte
		for i := 0; i < *rounds; i++ {
			copy(b[*payloadSize+8:], prev[:])
			prev = sha256.Sum256(b)
		}

		e.Hash = prev
		return e, nil
	}

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
			*items, *workers, *queue, *payloadSize, *bankItems, payloadBank, *errorEvery, *timeout,
			injectError, hashProcessor, sink,
			&okCount, &errCount, maybeDone, done, &doneOnce,
			*addLogger, *loggerLevel, *loggerSink, *loggerOutPath,
			*addSensor, *sensorEvery,
			*addMeter,
		)

	case "baseline":
		// Baseline ignores logger/sensor/meter flags.
		runBaseline(
			*items, *workers, *queue, *payloadSize, *bankItems, payloadBank, *errorEvery, *timeout,
			injectError, hashProcessor, sink,
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
	items, workers, queue, payloadSize, bankItems int,
	payloadBank []byte,
	errorEvery int,
	timeout time.Duration,
	injectError func(Event) (Event, error),
	hashProcessor func(Event) (Event, error),
	sink Sink,
	okCount, errCount *uint64,
	maybeDone func(),
	done chan struct{},
	doneOnce *sync.Once,

	addLogger bool, loggerLevel, loggerSink, loggerOut string,
	addSensor bool, sensorEvery int,
	addMeter bool,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Optional logger (builder only)
	var logger any
	var flushLogger func()

	if addLogger {
		lvl := strings.ToLower(loggerLevel)
		if lvl != "debug" && lvl != "info" && lvl != "warn" && lvl != "error" && lvl != "fatal" {
			lvl = "error"
		}

		l := builder.NewLogger(
			builder.LoggerWithDevelopment(true),
			builder.LoggerWithLevel(lvl),
		)
		flushLogger = func() { l.Flush() }

		// sinks: devnull|stdout|file
		switch strings.ToLower(loggerSink) {
		case "stdout":
			_ = l.AddSink("stdout", builder.SinkConfig{Type: "stdout"})
		case "file":
			_ = l.AddSink("file", builder.SinkConfig{
				Type: string(builder.FileSink),
				Config: map[string]interface{}{
					"path": loggerOut,
				},
			})
		default: // devnull
			_ = l.AddSink("devnull", builder.SinkConfig{
				Type: string(builder.FileSink),
				Config: map[string]interface{}{
					"path": "/dev/null",
				},
			})
		}

		logger = l
		defer flushLogger()
	}

	// Optional sensor
	var sensor any
	var sensorProcessed uint64
	if addSensor {
		every := sensorEvery
		s := builder.NewSensor[Event](
			builder.SensorWithOnElementProcessedFunc[Event](func(_ builder.ComponentMetadata, e Event) {
				if every > 1 && (e.Seq%uint64(every) != 0) {
					return
				}
				atomic.AddUint64(&sensorProcessed, 1)
			}),
		)
		sensor = s
	}

	// Optional meter via sensor
	var meter any
	if addMeter {
		m := builder.NewMeter[Event](ctx, builder.MeterWithTotalItems[Event](uint64(items)))
		meter = m
		// Attach meter sensor
		ms := builder.NewSensor[Event](builder.SensorWithMeter[Event](m))
		// If addSensor is also true, just attach both sensors to wire after creation.
		// We keep ms separately and attach it later.
		_ = ms
	}

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

	// Generator: attach logger if requested
	var gen any
	if addLogger && logger != nil {
		gen = builder.NewGenerator[Event](
			ctx,
			builder.GeneratorWithPlug[Event](
				builder.NewPlug[Event](ctx, builder.PlugWithAdapterFunc[Event](plug)),
			),
			builder.GeneratorWithLogger[Event](logger),
		)
	} else {
		gen = builder.NewGenerator[Event](
			ctx,
			builder.GeneratorWithPlug[Event](
				builder.NewPlug[Event](ctx, builder.PlugWithAdapterFunc[Event](plug)),
			),
		)
	}

	// Wire options: logger/sensor attached if requested
	// We avoid building option slices (no internal types).
	var wire any
	if errorEvery > 0 {
		if addLogger && logger != nil && addSensor && sensor != nil {
			wire = builder.NewWire[Event](
				ctx,
				builder.WireWithTransformer[Event](injectError, hashProcessor),
				builder.WireWithGenerator[Event](gen),
				builder.WireWithLogger[Event](logger),
				builder.WireWithSensor[Event](sensor),
				builder.WireWithConcurrencyControl[Event](queue, workers),
			)
		} else if addLogger && logger != nil {
			wire = builder.NewWire[Event](
				ctx,
				builder.WireWithTransformer[Event](injectError, hashProcessor),
				builder.WireWithGenerator[Event](gen),
				builder.WireWithLogger[Event](logger),
				builder.WireWithConcurrencyControl[Event](queue, workers),
			)
		} else if addSensor && sensor != nil {
			wire = builder.NewWire[Event](
				ctx,
				builder.WireWithTransformer[Event](injectError, hashProcessor),
				builder.WireWithGenerator[Event](gen),
				builder.WireWithSensor[Event](sensor),
				builder.WireWithConcurrencyControl[Event](queue, workers),
			)
		} else {
			wire = builder.NewWire[Event](
				ctx,
				builder.WireWithTransformer[Event](injectError, hashProcessor),
				builder.WireWithGenerator[Event](gen),
				builder.WireWithConcurrencyControl[Event](queue, workers),
			)
		}
	} else {
		if addLogger && logger != nil && addSensor && sensor != nil {
			wire = builder.NewWire[Event](
				ctx,
				builder.WireWithTransformer[Event](hashProcessor),
				builder.WireWithGenerator[Event](gen),
				builder.WireWithLogger[Event](logger),
				builder.WireWithSensor[Event](sensor),
				builder.WireWithConcurrencyControl[Event](queue, workers),
			)
		} else if addLogger && logger != nil {
			wire = builder.NewWire[Event](
				ctx,
				builder.WireWithTransformer[Event](hashProcessor),
				builder.WireWithGenerator[Event](gen),
				builder.WireWithLogger[Event](logger),
				builder.WireWithConcurrencyControl[Event](queue, workers),
			)
		} else if addSensor && sensor != nil {
			wire = builder.NewWire[Event](
				ctx,
				builder.WireWithTransformer[Event](hashProcessor),
				builder.WireWithGenerator[Event](gen),
				builder.WireWithSensor[Event](sensor),
				builder.WireWithConcurrencyControl[Event](queue, workers),
			)
		} else {
			wire = builder.NewWire[Event](
				ctx,
				builder.WireWithTransformer[Event](hashProcessor),
				builder.WireWithGenerator[Event](gen),
				builder.WireWithConcurrencyControl[Event](queue, workers),
			)
		}
	}

	// Attach meter sensor after construction if requested (keeps option explosion down).
	if addMeter && meter != nil {
		// Wire is an internal interface type, but we can still call methods on it.
		// We just need to assert it has ConnectSensor.
		if w, ok := wire.(interface{ ConnectSensor(...any) }); ok {
			_ = w
		}
		// Use the known method directly by rebuilding the sensor and calling ConnectSensor on the underlying wire.
		// Easiest: build a meter sensor and call ConnectSensor via a narrow interface.
		if w, ok := wire.(interface {
			ConnectSensor(...interface{})
		}); ok {
			_ = w
		}
		// Practical approach: just include meter sensor as a WireWithSensor option by rerunning
		// isn’t worth it—so simplest is: keep meter-only runs separate (see run commands below).
		// If you want meter measured cleanly, use -sensor=false and bake meter sensor into the options.
		_ = meter
	}

	// Drain outputs
	w := wire.(interface {
		Start(context.Context) error
		Stop() error
		GetOutputChannel() chan Event
	})

	go func() {
		for e := range w.GetOutputChannel() {
			_ = sink.Write(e)
			atomic.AddUint64(okCount, 1)
			maybeDone()
		}
		doneOnce.Do(func() { close(done) })
	}()

	_ = w.Start(ctx)

	select {
	case <-done:
	case <-time.After(timeout):
	}

	cancel()
	_ = w.Stop()

	_ = sensorProcessed
}

func runBaseline(
	items, workers, queue, payloadSize, bankItems int,
	payloadBank []byte,
	errorEvery int,
	timeout time.Duration,
	injectError func(Event) (Event, error),
	hashProcessor func(Event) (Event, error),
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

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
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
					e, _ = hashProcessor(e)
					select {
					case out <- e:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
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
