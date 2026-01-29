package s3client

import (
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SetS3ClientDeps wires the S3 client and bucket for the adapter.
func (a *S3Client[T]) SetS3ClientDeps(d types.S3ClientDeps) {
	a.cli = d.Client
	a.bucket = strings.TrimSpace(d.Bucket)
}

// SetWriterConfig applies writer configuration fields that are explicitly set.
func (a *S3Client[T]) SetWriterConfig(c types.S3WriterConfig) {
	if c.PrefixTemplate != "" {
		a.prefixTemplate = c.PrefixTemplate
	}
	if c.FileNameTmpl != "" {
		a.fileNameTmpl = c.FileNameTmpl
	}

	if c.Format != "" {
		a.formatName = strings.ToLower(c.Format)
	}

	if len(c.FormatOptions) > 0 {
		if a.formatOpts == nil {
			a.formatOpts = make(map[string]string, len(c.FormatOptions))
		}
		for k, v := range c.FormatOptions {
			a.formatOpts[k] = v
		}
	}

	if c.Compression != "" {
		a.ndjsonEncGz = strings.EqualFold(c.Compression, "gzip")
		if a.ndjsonEncGz {
			if a.formatOpts == nil {
				a.formatOpts = map[string]string{}
			}
			a.formatOpts["gzip"] = "true"
		}
	}

	a.sseMode, a.kmsKey = c.SSEMode, c.KMSKeyID
	if c.RequireSSE {
		a.requireSSE = true
	}

	if c.ClientSideEncryption != "" || c.ClientSideKey != "" || c.RequireClientSideEncryption {
		a.cseMode = strings.ToLower(strings.TrimSpace(c.ClientSideEncryption))
		if a.cseMode == "" {
			a.cseMode = cseModeAESGCM
		}
		if c.ClientSideKey != "" {
			key, err := parseAESGCMKeyHex(c.ClientSideKey)
			if err != nil {
				a.configErr = err
			} else {
				a.cseKey = key
			}
		}
		if c.RequireClientSideEncryption {
			a.requireCSE = true
		}
	}

	if c.BatchMaxRecords > 0 {
		a.batchMaxRecords = c.BatchMaxRecords
	}
	if c.BatchMaxBytes > 0 {
		a.batchMaxBytes = c.BatchMaxBytes
	}
	if c.BatchMaxAge > 0 {
		a.batchMaxAge = c.BatchMaxAge
	}

	if c.RawExtension != "" {
		a.rawWriterExt = c.RawExtension
	}
	if c.RawContentType != "" {
		a.rawWriterContentType = c.RawContentType
	}

	if strings.EqualFold(a.formatName, "parquet") {
		if a.rawWriterExt == "" {
			a.rawWriterExt = ".parquet"
		}
		if a.rawWriterContentType == "" {
			a.rawWriterContentType = "application/parquet"
		}
	}
}

// SetReaderConfig applies reader configuration fields that are explicitly set.
func (a *S3Client[T]) SetReaderConfig(c types.S3ReaderConfig) {
	a.listPrefix = c.Prefix
	a.listStartAfter = c.StartAfterKey
	if c.PageSize > 0 {
		a.listPageSize = c.PageSize
	}
	if c.ListInterval > 0 {
		a.listPollInterval = c.ListInterval
	}

	if c.Format != "" {
		a.readerFormatName = strings.ToLower(c.Format)
	}
	if len(c.FormatOptions) > 0 {
		if a.readerFormatOpts == nil {
			a.readerFormatOpts = make(map[string]string, len(c.FormatOptions))
		}
		for k, v := range c.FormatOptions {
			a.readerFormatOpts[k] = v
		}
	}
	if c.Compression != "" && strings.EqualFold(c.Compression, "gzip") {
		if a.readerFormatOpts == nil {
			a.readerFormatOpts = map[string]string{}
		}
		a.readerFormatOpts["gzip"] = "true"
	}

	if c.ClientSideEncryption != "" || c.ClientSideKey != "" || c.RequireClientSideEncryption {
		a.cseMode = strings.ToLower(strings.TrimSpace(c.ClientSideEncryption))
		if a.cseMode == "" {
			a.cseMode = cseModeAESGCM
		}
		if c.ClientSideKey != "" {
			key, err := parseAESGCMKeyHex(c.ClientSideKey)
			if err != nil {
				a.configErr = err
			} else {
				a.cseKey = key
			}
		}
		if c.RequireClientSideEncryption {
			a.requireCSE = true
		}
	}
}
