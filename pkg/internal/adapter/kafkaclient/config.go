package kafkaclient

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// SetKafkaClientDeps wires broker addresses and driver-specific dependencies.
func (a *KafkaClient[T]) SetKafkaClientDeps(d types.KafkaClientDeps) {
	a.brokers = append([]string(nil), d.Brokers...)
	a.producer = d.Producer
	a.consumer = d.Consumer
	a.dlqTopic = d.DLQTopic
}

// SetWriterConfig applies writer configuration fields that are explicitly set.
func (a *KafkaClient[T]) SetWriterConfig(c types.KafkaWriterConfig) {
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

	if c.BatchMaxRecords > 0 {
		a.wBatchMaxRecords = c.BatchMaxRecords
	}
	if c.BatchMaxBytes > 0 {
		a.wBatchMaxBytes = c.BatchMaxBytes
	}
	if c.BatchMaxAge > 0 {
		a.wBatchMaxAge = c.BatchMaxAge
	}

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

// SetReaderConfig applies reader configuration fields that are explicitly set.
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
