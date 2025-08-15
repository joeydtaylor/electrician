package kafkaclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	kafka "github.com/segmentio/kafka-go"
)

// ---------- deps & config setters ----------

func (a *KafkaClient[T]) SetKafkaClientDeps(d types.KafkaClientDeps) {
	a.brokers = append([]string(nil), d.Brokers...)
	a.producer = d.Producer
	a.consumer = d.Consumer
	a.dlqTopic = d.DLQTopic
}

func (a *KafkaClient[T]) SetWriterConfig(c types.KafkaWriterConfig) {
	// destination / format
	if c.Topic != "" {
		a.wTopic = c.Topic
	}
	if c.Format != "" {
		a.wFormat = c.Format
	}
	if len(c.FormatOptions) > 0 {
		if a.wFormatOpts == nil {
			a.wFormatOpts = make(map[string]string, len(c.FormatOptions))
		}
		for k, v := range c.FormatOptions {
			a.wFormatOpts[k] = v
		}
	}
	if c.Compression != "" {
		a.wCompression = c.Compression
	}
	if c.KeyTemplate != "" {
		a.wKeyTemplate = c.KeyTemplate
	}
	if len(c.HeaderTemplates) > 0 {
		if a.wHdrTemplates == nil {
			a.wHdrTemplates = make(map[string]string, len(c.HeaderTemplates))
		}
		for k, v := range c.HeaderTemplates {
			a.wHdrTemplates[k] = v
		}
	}

	// batching
	if c.BatchMaxRecords > 0 {
		a.wBatchMaxRecords = c.BatchMaxRecords
	}
	if c.BatchMaxBytes > 0 {
		a.wBatchMaxBytes = c.BatchMaxBytes
	}
	if c.BatchMaxAge > 0 {
		a.wBatchMaxAge = c.BatchMaxAge
	}

	// producer semantics
	if c.Acks != "" {
		a.wAcks = c.Acks
	}
	if c.RequestTimeout > 0 {
		a.wReqTimeout = c.RequestTimeout
	}
	if c.PartitionStrategy != "" {
		a.wPartitionStrat = c.PartitionStrategy
	}
	a.wManualPartition = c.ManualPartition
	a.wEnableDLQ = c.EnableDLQ
}

func (a *KafkaClient[T]) SetReaderConfig(c types.KafkaReaderConfig) {
	if c.GroupID != "" {
		a.rGroupID = c.GroupID
	}
	if len(c.Topics) > 0 {
		a.rTopics = append([]string(nil), c.Topics...)
	}
	if c.StartAt != "" {
		a.rStartAt = c.StartAt
	}
	if !c.StartAtTime.IsZero() {
		a.rStartAtTime = c.StartAtTime
	}
	if c.PollInterval > 0 {
		a.rPollInterval = c.PollInterval
	}
	if c.MaxPollRecords > 0 {
		a.rMaxPollRecs = c.MaxPollRecords
	}
	if c.MaxPollBytes > 0 {
		a.rMaxPollBytes = c.MaxPollBytes
	}
	if c.Format != "" {
		a.rFormat = c.Format
	}
	if len(c.FormatOptions) > 0 {
		if a.rFormatOpts == nil {
			a.rFormatOpts = make(map[string]string, len(c.FormatOptions))
		}
		for k, v := range c.FormatOptions {
			a.rFormatOpts[k] = v
		}
	}
	if c.CommitMode != "" {
		a.rCommitMode = c.CommitMode
	}
	if c.CommitPolicy != "" {
		a.rCommitPolicy = c.CommitPolicy
	}
	if c.CommitInterval > 0 {
		a.rCommitEvery = c.CommitInterval
	}
}

// ---------- plumbing (loggers/sensors/metadata) ----------

func (a *KafkaClient[T]) ConnectSensor(s ...types.Sensor[T]) {
	a.sensorLock.Lock()
	local := make([]types.Sensor[T], 0, len(a.sensors)+len(s))
	local = append(local, a.sensors...)
	local = append(local, s...)
	a.sensors = local
	a.sensorLock.Unlock()

	for _, m := range s {
		a.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, event: ConnectSensor, target: %v",
			a.componentMetadata, m.GetComponentMetadata())
	}
}

func (a *KafkaClient[T]) ConnectLogger(l ...types.Logger) {
	a.loggersLock.Lock()
	a.loggers = append(a.loggers, l...)
	a.loggersLock.Unlock()
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ConnectLogger, total_loggers: %d",
		a.componentMetadata, len(a.loggers))
}

func (a *KafkaClient[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if len(a.loggers) == 0 {
		return
	}
	msg := fmt.Sprintf(format, args...)
	a.loggersLock.Lock()
	defer a.loggersLock.Unlock()
	for _, logger := range a.loggers {
		if logger == nil || logger.GetLevel() > level {
			continue
		}
		switch level {
		case types.DebugLevel:
			logger.Debug(msg)
		case types.InfoLevel:
			logger.Info(msg)
		case types.WarnLevel:
			logger.Warn(msg)
		case types.ErrorLevel:
			logger.Error(msg)
		case types.DPanicLevel:
			logger.DPanic(msg)
		case types.PanicLevel:
			logger.Panic(msg)
		case types.FatalLevel:
			logger.Fatal(msg)
		}
	}
}

func (a *KafkaClient[T]) GetComponentMetadata() types.ComponentMetadata { return a.componentMetadata }

func (a *KafkaClient[T]) SetComponentMetadata(name, id string) {
	a.componentMetadata = types.ComponentMetadata{Name: name, ID: id}
}

func (a *KafkaClient[T]) Name() string { return "KAFKA_CLIENT" }

// Stop stops both writer and reader sides and emits sensor stop hooks.
func (a *KafkaClient[T]) Stop() {
	a.forEachSensor(func(s types.Sensor[T]) {
		// Writer stop
		s.InvokeOnKafkaWriterStop(a.componentMetadata)
		// Consumer stop
		s.InvokeOnKafkaConsumerStop(a.componentMetadata)
	})
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Stop", a.componentMetadata)
	a.cancel()
}

// ---------- writer side ----------

func (a *KafkaClient[T]) ConnectInput(ws ...types.Wire[T]) {
	if len(ws) == 0 {
		return
	}
	a.inputWires = append(a.inputWires, ws...)
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ConnectInput, wires_added: %d, total_wires: %d",
		a.componentMetadata, len(ws), len(a.inputWires))
}

func (a *KafkaClient[T]) StartWriter(ctx context.Context) error {
	if a.wTopic == "" {
		return fmt.Errorf("kafkaclient: StartWriter requires Topic; call SetWriterConfig(...)")
	}
	if len(a.inputWires) == 0 {
		return fmt.Errorf("kafkaclient: StartWriter requires at least one connected wire; call ConnectInput(...)")
	}

	// allocate merged channel based on batch size
	if a.mergedIn == nil {
		size := a.wBatchMaxRecords
		if size <= 0 {
			size = 1024
		}
		a.mergedIn = make(chan T, size)
	}

	// fan-in all wires
	for _, w := range a.inputWires {
		if w == nil {
			continue
		}
		out := w.GetOutputChannel()
		go a.fanIn(ctx, a.mergedIn, out)
	}

	// delegate the serving to ServeWriter (single source of truth)
	go func() {
		_ = a.ServeWriter(ctx, a.mergedIn)
	}()
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: StartWriter, topic: %s, format: %s, wires: %d",
		a.componentMetadata, a.wTopic, a.wFormat, len(a.inputWires))

	return nil
}

// ServeWriter consumes records from `in` and produces to Kafka (record-oriented "ndjson").
func (a *KafkaClient[T]) ServeWriter(ctx context.Context, in <-chan T) error {
	effTopic, ok := a.effectiveWriterTopic()
	if !ok {
		return fmt.Errorf("kafkaclient: ServeWriter requires topic (set KafkaWriterConfig.Topic or use a kafka-go Writer with Topic)")
	}

	format := strings.ToLower(strings.TrimSpace(a.wFormat))
	if format == "" {
		format = "ndjson"
	}
	if format != "ndjson" {
		return fmt.Errorf("kafkaclient: ServeWriter only supports ndjson (got %q)", format)
	}

	if !atomic.CompareAndSwapInt32(&a.isServingWriter, 0, 1) {
		return nil // already running
	}
	defer atomic.StoreInt32(&a.isServingWriter, 0)

	// sensors + logger start/stop
	a.forEachSensor(func(s types.Sensor[T]) {
		s.InvokeOnKafkaWriterStart(a.componentMetadata, effTopic, format)
	})
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: WriterStart, topic: %s, format: %s",
		a.componentMetadata, effTopic, format)
	defer func() {
		a.forEachSensor(func(s types.Sensor[T]) { s.InvokeOnKafkaWriterStop(a.componentMetadata) })
		a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: WriterStop", a.componentMetadata)
	}()

	// batching defaults
	maxRecs := a.wBatchMaxRecords
	if maxRecs <= 0 {
		maxRecs = 1000
	}
	maxBytes := a.wBatchMaxBytes
	if maxBytes <= 0 {
		maxBytes = 1 << 20 // 1 MiB
	}
	maxAge := a.wBatchMaxAge
	if maxAge <= 0 {
		maxAge = time.Second
	}

	type msg struct {
		key     []byte
		val     []byte
		headers []struct{ Key, Value string }
	}
	pending := make([]msg, 0, maxRecs)
	var byteTally int

	tick := time.NewTicker(maxAge)
	defer tick.Stop()

	flush := func(_ time.Time) error {
		if len(pending) == 0 {
			return nil
		}
		// batch flush hook + logger
		a.forEachSensor(func(s types.Sensor[T]) {
			s.InvokeOnKafkaBatchFlush(a.componentMetadata, effTopic, len(pending), byteTally, a.wCompression)
		})
		a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: BatchFlush, topic: %s, records: %d, bytes: %d, compression: %s",
			a.componentMetadata, effTopic, len(pending), byteTally, a.wCompression)

		partForHook := -1
		if a.wManualPartition != nil {
			partForHook = *a.wManualPartition
		}

		for _, m := range pending {
			// attempt hook (sensor already emits); add a debug log too
			a.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, event: ProduceAttempt, topic: %s, partition: %d, key_bytes: %d, val_bytes: %d",
				a.componentMetadata, effTopic, partForHook, len(m.key), len(m.val))

			// pass "" as topic when the injected writer has a Topic set
			topicForMessage := ""
			if _, has := a.producer.(*kafka.Writer); !has {
				topicForMessage = effTopic
			}

			partition, offset, err := a.produce(ctx, topicForMessage, a.wManualPartition, m.key, m.val, m.headers)
			if err != nil {
				a.forEachSensor(func(s types.Sensor[T]) {
					s.InvokeOnKafkaProduceError(a.componentMetadata, effTopic, partForHook, err)
				})
				a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Produce, topic: %s, err: %v",
					a.componentMetadata, effTopic, err)
				return err
			}

			// success hook + logger
			a.forEachSensor(func(s types.Sensor[T]) {
				s.InvokeOnKafkaProduceSuccess(a.componentMetadata, effTopic, partition, offset, 0 /*dur unknown*/)
			})
			offsetStr := "n/a"
			if offset >= 0 {
				offsetStr = fmt.Sprintf("%d", offset)
			}
			a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ProduceSuccess, topic: %s, partition: %d, offset: %s",
				a.componentMetadata, effTopic, partition, offsetStr)
		}

		pending = pending[:0]
		byteTally = 0
		return nil
	}

	encode := func(v T) (key, val []byte, hdrs []struct{ Key, Value string }, err error) {
		// value as NDJSON
		val, err = json.Marshal(v)
		if err != nil {
			return nil, nil, nil, err
		}
		val = append(val, '\n')

		// key/headers via simple templates
		key = renderKeyFromTemplate(a.wKeyTemplate, v)
		hdrs = renderHeadersFromTemplates(a.wHdrTemplates, v)

		// hooks for adornments + debug logs
		a.forEachSensor(func(s types.Sensor[T]) {
			if key != nil {
				s.InvokeOnKafkaKeyRendered(a.componentMetadata, key)
			}
			if hdrs != nil {
				s.InvokeOnKafkaHeadersRendered(a.componentMetadata, hdrs)
			}
		})
		if key != nil {
			a.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, event: KeyRendered, bytes: %d", a.componentMetadata, len(key))
		}
		if hdrs != nil {
			a.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, event: HeadersRendered, count: %d", a.componentMetadata, len(hdrs))
		}
		return key, val, hdrs, nil
	}

	for {
		select {
		case <-ctx.Done():
			_ = flush(time.Now())
			return nil

		case m, ok := <-in:
			if !ok {
				_ = flush(time.Now())
				return nil
			}
			k, v, h, err := encode(m)
			if err != nil {
				a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Encode, err: %v", a.componentMetadata, err)
				continue
			}
			pending = append(pending, msg{key: k, val: v, headers: h})
			byteTally += len(k) + len(v)

			if len(pending) >= maxRecs || byteTally >= maxBytes {
				if err := flush(time.Now()); err != nil {
					return err
				}
			}

		case <-tick.C:
			if err := flush(time.Now()); err != nil {
				return err
			}
		}
	}
}

func (a *KafkaClient[T]) ServeWriterRaw(ctx context.Context, in <-chan []byte) error {
	effTopic, ok := a.effectiveWriterTopic()
	if !ok {
		return fmt.Errorf("kafkaclient: ServeWriterRaw requires topic (set KafkaWriterConfig.Topic or use a kafka-go Writer with Topic)")
	}

	if !atomic.CompareAndSwapInt32(&a.isServingWriter, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&a.isServingWriter, 0)

	// start/stop hooks + logs
	a.forEachSensor(func(s types.Sensor[T]) {
		s.InvokeOnKafkaWriterStart(a.componentMetadata, effTopic, "raw")
	})
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: WriterStart, topic: %s, format: raw",
		a.componentMetadata, effTopic)
	defer func() {
		a.forEachSensor(func(s types.Sensor[T]) { s.InvokeOnKafkaWriterStop(a.componentMetadata) })
		a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: WriterStop", a.componentMetadata)
	}()

	// batching defaults
	maxRecs := a.wBatchMaxRecords
	if maxRecs <= 0 {
		maxRecs = 1000
	}
	maxBytes := a.wBatchMaxBytes
	if maxBytes <= 0 {
		maxBytes = 1 << 20 // 1 MiB
	}
	maxAge := a.wBatchMaxAge
	if maxAge <= 0 {
		maxAge = time.Second
	}

	type msg struct{ val []byte }
	pending := make([]msg, 0, maxRecs)
	var byteTally int

	tick := time.NewTicker(maxAge)
	defer tick.Stop()

	flush := func(_ time.Time) error {
		if len(pending) == 0 {
			return nil
		}
		// batch flush
		a.forEachSensor(func(s types.Sensor[T]) {
			s.InvokeOnKafkaBatchFlush(a.componentMetadata, effTopic, len(pending), byteTally, a.wCompression)
		})
		a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: BatchFlush, topic: %s, records: %d, bytes: %d, compression: %s",
			a.componentMetadata, effTopic, len(pending), byteTally, a.wCompression)

		partForHook := -1
		if a.wManualPartition != nil {
			partForHook = *a.wManualPartition
		}

		for _, m := range pending {
			a.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, event: ProduceAttempt, topic: %s, partition: %d, key_bytes: %d, val_bytes: %d",
				a.componentMetadata, effTopic, partForHook, 0, len(m.val))

			// pass "" as topic when the injected writer has a Topic set
			topicForMessage := ""
			if _, has := a.producer.(*kafka.Writer); !has {
				topicForMessage = effTopic
			}

			partition, offset, err := a.produce(ctx, topicForMessage, a.wManualPartition, nil, m.val, nil)
			if err != nil {
				a.forEachSensor(func(s types.Sensor[T]) {
					s.InvokeOnKafkaProduceError(a.componentMetadata, effTopic, partForHook, err)
				})
				a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Produce, topic: %s, err: %v",
					a.componentMetadata, effTopic, err)
				return err
			}

			a.forEachSensor(func(s types.Sensor[T]) {
				s.InvokeOnKafkaProduceSuccess(a.componentMetadata, effTopic, partition, offset, 0)
			})
			offsetStr := "n/a"
			if offset >= 0 {
				offsetStr = fmt.Sprintf("%d", offset)
			}
			a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ProduceSuccess, topic: %s, partition: %d, offset: %s",
				a.componentMetadata, effTopic, partition, offsetStr)
		}

		pending = pending[:0]
		byteTally = 0
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			_ = flush(time.Now())
			return nil
		case b, ok := <-in:
			if !ok {
				_ = flush(time.Now())
				return nil
			}
			if len(b) == 0 {
				continue
			}
			pending = append(pending, msg{val: append([]byte(nil), b...)})
			byteTally += len(b)

			if len(pending) >= maxRecs || byteTally >= maxBytes {
				if err := flush(time.Now()); err != nil {
					return err
				}
			}
		case <-tick.C:
			if err := flush(time.Now()); err != nil {
				return err
			}
		}
	}
}

// ---------- reader side ----------

func (a *KafkaClient[T]) Serve(ctx context.Context, submit func(context.Context, T) error) error {
	if len(a.rTopics) == 0 {
		return fmt.Errorf("kafkaclient: Serve requires at least one Topic; call SetReaderConfig(...)")
	}
	if !atomic.CompareAndSwapInt32(&a.isServingReader, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&a.isServingReader, 0)

	r, created, err := a.getOrCreateReaderForServe()
	if err != nil {
		return err
	}
	if created {
		defer func() { _ = r.Close() }()
		a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ReaderCreated, group: %s, topics: %v",
			a.componentMetadata, a.rGroupID, a.rTopics)
	}

	// normalize commit policy aliases
	policy := strings.ToLower(strings.TrimSpace(a.rCommitPolicy))
	if policy == "interval" {
		policy = "time"
	}
	mode := strings.ToLower(strings.TrimSpace(a.rCommitMode))
	if mode == "" {
		mode = "auto"
	}

	// emit consumer start
	a.forEachSensor(func(s types.Sensor[T]) {
		s.InvokeOnKafkaConsumerStart(a.componentMetadata, a.rGroupID, append([]string(nil), a.rTopics...), a.rStartAt)
	})
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ConsumerStart, group: %s, topics: %v, start_at: %s",
		a.componentMetadata, a.rGroupID, a.rTopics, a.rStartAt)
	defer func() {
		a.forEachSensor(func(s types.Sensor[T]) { s.InvokeOnKafkaConsumerStop(a.componentMetadata) })
		a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ConsumerStop", a.componentMetadata)
	}()

	maxPollRecs := a.rMaxPollRecs
	if maxPollRecs <= 0 {
		maxPollRecs = 10_000
	}
	maxPollBytes := a.rMaxPollBytes
	if maxPollBytes <= 0 {
		maxPollBytes = 1 << 20 // 1 MiB
	}
	pollEvery := a.rPollInterval
	if pollEvery <= 0 {
		pollEvery = 1 * time.Second
	}

	// manual commit machinery
	latestByTP := make(map[string]kafka.Message) // key "topic:partition"
	var commitTicker *time.Ticker
	if mode == "manual" && policy == "time" && a.rCommitEvery > 0 {
		commitTicker = time.NewTicker(a.rCommitEvery)
		defer commitTicker.Stop()
	}

	for {
		// allow graceful exit at a bounded cadence
		windowCtx, cancel := context.WithTimeout(ctx, pollEvery)

		decodedByTopic := map[string]int{} // for OnKafkaDecode
		fetched := 0
		byteSum := 0

	inner:
		for fetched < maxPollRecs && byteSum < maxPollBytes {
			msg, err := r.FetchMessage(windowCtx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					break inner
				}
				a.NotifyLoggers(types.WarnLevel, "%s => level: WARN, event: FetchMessage, err: %v", a.componentMetadata, err)
				break inner
			}

			// message hook + logger
			a.forEachSensor(func(s types.Sensor[T]) {
				s.InvokeOnKafkaMessage(a.componentMetadata, msg.Topic, msg.Partition, msg.Offset, len(msg.Key), len(msg.Value))
			})
			a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Message, topic: %s, partition: %d, offset: %d, key_bytes: %d, val_bytes: %d",
				a.componentMetadata, msg.Topic, msg.Partition, msg.Offset, len(msg.Key), len(msg.Value))

			// decode
			var v T
			switch strings.ToLower(strings.TrimSpace(a.rFormat)) {
			case "ndjson", "", "json":
				if err := json.Unmarshal(msg.Value, &v); err != nil {
					a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Decode, topic: %s, err: %v", a.componentMetadata, msg.Topic, err)
					continue
				}
			default:
				// support for additional decoders can be added here; default to JSON
				if err := json.Unmarshal(msg.Value, &v); err != nil {
					a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Decode, topic: %s, err: %v", a.componentMetadata, msg.Topic, err)
					continue
				}
			}
			decodedByTopic[msg.Topic]++

			// deliver to caller
			if err := submit(ctx, v); err != nil {
				a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Submit, err: %v", a.componentMetadata, err)
				// do not commit this one on failure
				continue
			}

			// commit strategy
			switch mode {
			case "manual":
				switch policy {
				case "after-each":
					if err := r.CommitMessages(ctx, msg); err != nil {
						a.forEachSensor(func(s types.Sensor[T]) { s.InvokeOnKafkaCommitError(a.componentMetadata, a.rGroupID, err) })
						a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Commit(after-each), err: %v", a.componentMetadata, err)
					} else {
						key := fmt.Sprintf("%s:%d", msg.Topic, msg.Partition)
						a.forEachSensor(func(s types.Sensor[T]) {
							s.InvokeOnKafkaCommitSuccess(a.componentMetadata, a.rGroupID, map[string]int64{key: msg.Offset + 1})
						})
						a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: CommitSuccess(after-each), group: %s, offsets: %s=%d",
							a.componentMetadata, a.rGroupID, key, msg.Offset+1)
					}
				case "after-batch", "time", "":
					key := fmt.Sprintf("%s:%d", msg.Topic, msg.Partition)
					latestByTP[key] = msg // keep latest per partition
				default:
					key := fmt.Sprintf("%s:%d", msg.Topic, msg.Partition)
					latestByTP[key] = msg
				}
			case "auto":
				// kafka-go will auto-commit at ReaderConfig.CommitInterval
			}

			fetched++
			byteSum += len(msg.Key) + len(msg.Value)
		}

		cancel() // end of this polling window

		// decode hooks per topic + log
		if len(decodedByTopic) > 0 {
			for t, n := range decodedByTopic {
				a.forEachSensor(func(s types.Sensor[T]) {
					s.InvokeOnKafkaDecode(a.componentMetadata, t, n, a.rFormat)
				})
				a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Decode, topic: %s, rows: %d, format: %s",
					a.componentMetadata, t, n, a.rFormat)
			}
		}

		// commit after-batch
		if mode == "manual" && policy == "after-batch" && len(latestByTP) > 0 {
			a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Commit(after-batch), partitions: %d",
				a.componentMetadata, len(latestByTP))
			a.commitLatest(ctx, r, latestByTP, a.rGroupID)
			latestByTP = make(map[string]kafka.Message)
		}

		// time-based commit tick (non-blocking)
		if commitTicker != nil {
			select {
			case <-commitTicker.C:
				if len(latestByTP) > 0 {
					a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Commit(time), partitions: %d, every: %s",
						a.componentMetadata, len(latestByTP), a.rCommitEvery)
					a.commitLatest(ctx, r, latestByTP, a.rGroupID)
					latestByTP = make(map[string]kafka.Message)
				}
			default:
			}
		}

		// overall exit?
		select {
		case <-ctx.Done():
			// best-effort final commit if manual
			if mode == "manual" && len(latestByTP) > 0 {
				a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Commit(final), partitions: %d",
					a.componentMetadata, len(latestByTP))
				a.commitLatest(context.Background(), r, latestByTP, a.rGroupID)
			}
			return nil
		default:
		}
	}
}

func (a *KafkaClient[T]) Fetch() (types.HttpResponse[[]T], error) {
	// Optional one-shot helper — not implemented by design
	return types.HttpResponse[[]T]{StatusCode: 501, Body: nil}, fmt.Errorf("kafkaclient: Fetch not implemented")
}
