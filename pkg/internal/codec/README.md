# 🎛️ Codec Package

The `codec` package is Electrician’s **encode/decode toolbox**.

It provides small, composable encoders/decoders used at system boundaries:

* inbound ingestion (HTTP server / adapters)
* outbound emission (relays / sinks)
* internal buffering/serialization (optional wire encoder)

Codecs are intentionally simple: they turn **bytes ↔ typed values** using `io.Reader` / `io.Writer`.

---

## ✅ What it does

| Capability                      | Meaning                                                                            |
| ------------------------------- | ---------------------------------------------------------------------------------- |
| 🔁 Encode / decode              | Convert between `T` and bytes via streaming interfaces.                            |
| 🧩 Small format implementations | JSON/XML/text/line/binary/html helpers (depending on what you import).             |
| 🧬 Typed contracts              | Aligns with Electrician’s `types.Encoder[T]` / `types.Decoder[T]` style contracts. |
| 📦 Boundary glue                | Makes it easy for adapters/relays to share the same serialization behavior.        |

---

## ❌ What it isn’t

* A universal serialization framework.
* A guarantee that every format is “type-safe” in the strong sense (JSON/XML still depend on your struct tags and schema discipline).
* A compression system for transport (compression choices are usually part of relays, parquet, or adapter implementations).

---

## 📂 Package layout

This is the conceptual layout based on the current file split:

| File            | Responsibility                                                                 |
| --------------- | ------------------------------------------------------------------------------ |
| `codec.go`      | Core codec interface(s) and shared contracts                                   |
| `api.go`        | Constructors/helpers for building codecs (and any registry helpers if present) |
| `json.go`       | JSON encode/decode helpers                                                     |
| `xml.go`        | XML encode/decode helpers                                                      |
| `text.go`       | Plain text encode/decode helpers                                               |
| `line.go`       | Line-oriented encoding/decoding (newline-delimited patterns)                   |
| `binary.go`     | Binary encoding/decoding helpers (raw bytes, fixed layouts, etc.)              |
| `html.go`       | HTML parsing helpers                                                           |
| `wave.go`       | Waveform/signal encoding helpers (specialized)                                 |
| `codec_test.go` | Tests                                                                          |

---

## 🧠 Dependency posture (accurate)

Most codecs lean on the Go standard library (`encoding/json`, `encoding/xml`, `bufio`, `io`, etc.).

There are two common reasons you’ll see non-stdlib deps here:

* **HTML parsing**: Go’s HTML parser lives under `golang.org/x/net/html` (not the core stdlib).
* **Wave/DSP utilities**: waveform/signal helpers may use math/DSP libraries (e.g., `gonum` / `go-dsp`) when you opt into those features.

If you never import the HTML or wave code paths, those dependencies won’t matter to your final binary.

---

## 🔧 How codecs are used

Typical patterns:

* **Decode on ingress**: `io.Reader` → `T` (HTTP body, Kafka message, file input)
* **Encode on egress**: `T` → `io.Writer` (relay payload, output buffer, file sink)

Keep heavy transforms out of codecs. Codecs should do serialization/deserialization only; business logic belongs in transformers.

---

## 🔧 Extending the codec package

When adding a new format:

1. Implement an encoder and/or decoder using `io.Reader` / `io.Writer`.
2. Keep the API boring: explicit constructors, explicit configuration.
3. Add tests that cover:

   * correct round-trip behavior where applicable
   * error handling on invalid input
   * no accidental allocations/regressions on hot paths (where relevant)

Only update `types/` if multiple packages need a new shared contract; otherwise keep changes local to `codec`.

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.md)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/wave_encoding_example/)** – Demonstrates **Codecs in action**.

---

## 📝 License

The **Codec package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy encoding! ⚡📦** If you have questions or need support, feel free to open a GitHub issue.
