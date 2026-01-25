# codec

The codec package provides encoding and decoding helpers for pipeline payloads. It includes common formats and compression paths used by adapters and relays.

## Responsibilities

- Encode and decode structured payloads (JSON, XML, text, line-delimited).
- Support binary and wave helpers where applicable.
- Provide uniform interfaces for adapters and relays.

## Key files

- json.go, xml.go, text.go, html.go, line.go: text-based codecs
- binary.go: binary payload helpers
- wave.go: waveform encoding helpers

## Usage

Codecs are typically used by adapters or relays rather than by user code directly. The builder package re-exports or wraps codec utilities when needed.

## References

- builder: pkg/builder/codec.go
- internal contracts: pkg/internal/types/codec.go
