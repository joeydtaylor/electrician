# 🎛️ Codec Package

The **Codec** package provides **efficient, type-safe encoding and decoding**  
for **JSON, XML, Text, Binary, HTML, and custom formats**.

It enables **seamless data transformation**, ensuring **reliable serialization**  
for **network communication, storage, and structured logging**.

---

## 📦 Package Overview

| Feature                    | Description                                                        |
| -------------------------- | ------------------------------------------------------------------ |
| **JSON, XML, and Text**    | Supports **structured serialization and deserialization**.         |
| **Binary & Waveform Data** | Encodes **raw bytes and complex waveforms** for signal processing. |
| **HTML Parsing**           | Extracts **HTML nodes and document structure**.                    |
| **Streaming Support**      | Works with **io.Reader and io.Writer** for efficient I/O handling. |

---

## 📂 Package Structure

| File              | Purpose                                                            |
| ----------------- | ------------------------------------------------------------------ |
| **api.go**        | Public API for **codec registration and usage**.                   |
| **binary.go**     | Handles **binary encoding and decoding**.                          |
| **codec.go**      | Core **Codec interface** defining common serialization methods.    |
| **html.go**       | Parses **HTML documents into structured node representations**.    |
| **json.go**       | Implements **JSON encoding/decoding for generic types**.           |
| **line.go**       | Provides **line-based encoding** for text-based formats.           |
| **text.go**       | Handles **plain text serialization and deserialization**.          |
| **wave.go**       | Encodes and decodes **waveform data** with **frequency analysis**. |
| **xml.go**        | Implements **XML serialization and deserialization**.              |
| **codec_test.go** | Unit tests ensuring **correctness, efficiency, and reliability**.  |

---

## ⚡ Notable Dependencies

Electrician is **built primarily on Go’s standard library**, with two notable exceptions:

1. **Logging:** Uses Zap, which is the fastest structured logger for Go.
2. **Compression & Encoding (Protobuf Relay Only):**  
   Uses widely adopted **ZSTD, Snappy, Brotli, LZ4, and Deflate** for optimized performance.

For **all other encoding formats** (JSON, XML, Text, Binary, and HTML),  
Electrician **relies solely on Go’s standard library** for **maximum compatibility and efficiency**.

---

## 🔧 How Codecs Work

A **Codec** provides a **unified interface** for **serializing and deserializing** data in multiple formats.

### ✅ **Key Mechanisms**

- **Generic Serialization:** Supports **any Go type** via **generic encoders and decoders**.
- **Stream-Based Processing:** Works with **io.Reader and io.Writer** for **efficient data handling**.
- **Binary Encoding:** Supports **custom binary structures** like **Waveforms and Frequency Peaks**.
- **HTML Decoding:** Extracts structured **HTML nodes for parsing and analysis**.

---

## 🔧 Extending the Codec Package

To **add new encoding formats**, follow this **structured workflow**:

### 1️⃣ Implement a New Encoder & Decoder

- Create a new file inside the codec package.
- Implement **Encoder[T]** and **Decoder[T]** interfaces.

### 2️⃣ Register in `api.go`

- Add the new encoder and decoder **to the codec registry**.

### 3️⃣ Unit Testing (`codec_test.go`)

- Ensure the **new format is tested under real-world conditions**.

---

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
