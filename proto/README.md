# ⚡ Electrician Protocol Buffers (Protobuf)

The **Electrician Protocol Buffers** define the core message formats and services that enable structured, efficient communication across the Electrician ecosystem. These definitions ensure **type safety**, **compression efficiency**, and **seamless data exchange** between components, whether in a local pipeline or a distributed microservice architecture.

---

## 📋 Protobuf Definitions

```protobuf
syntax = "proto3";

import "google/protobuf/timestamp.proto";

package electrician;

option go_package = "pkg/internal/relay";

message WrappedPayload {
  string id = 1;
  google.protobuf.Timestamp timestamp = 2;
  bytes payload = 3;
  MessageMetadata metadata = 4;
  ErrorInfo error_info = 5;
}

message ErrorInfo {
  int32 code = 1;
  string message = 2;
  repeated string details = 3; // More detailed error descriptions.
}

message MessageMetadata {
  map<string, string> headers = 1;
  string content_type = 2;
  VersionInfo version = 3;
  PerformanceOptions performance = 4;
  string trace_id = 5; // For distributed tracing.
  int32 priority = 6; // Priority of the message, higher number indicates higher priority.
}

message VersionInfo {
  int32 major = 1;
  int32 minor = 2;
}

message PerformanceOptions {
  bool use_compression = 1;
  CompressionAlgorithm compression_algorithm = 2;
  int32 compression_level = 3; // Optional field for specifying compression level.
}

enum CompressionAlgorithm {
  COMPRESS_NONE = 0;
  COMPRESS_DEFLATE = 1;
  COMPRESS_SNAPPY = 2;
  COMPRESS_ZSTD = 3;  // Zstandard, good balance of speed and compression ratio
  COMPRESS_BROTLI = 4; // Brotli, optimized for text compression
  COMPRESS_LZ4 = 5;    // LZ4, extremely fast compression and decompression
}

service RelayService {
  rpc Receive(WrappedPayload) returns (StreamAcknowledgment);
  rpc StreamReceive(stream WrappedPayload) returns (stream StreamAcknowledgment);
}

message StreamAcknowledgment {
  bool success = 1;
  string message = 2;
  map<string, string> metadata = 3; // Additional metadata about the acknowledgment.
}
```

## 📌 Message Structure & Components

### 🔹 WrappedPayload

Represents a structured data payload containing:

- A unique **`id`**
- A **`timestamp`** for event ordering
- The **`payload`** (actual data)
- **`MessageMetadata`** for additional context (headers, tracing, priority, etc.)
- **`ErrorInfo`** for error reporting

### 🔹 ErrorInfo

Encapsulates error details, including:

- A status **`code`**
- A descriptive **`message`**
- A **`details`** field for additional debugging information

### 🔹 MessageMetadata

Carries contextual metadata for each message:

- Custom **`headers`**
- **`content_type`** for structured data handling
- **`trace_id`** for distributed tracing
- **`priority`** field to indicate message importance
- **`VersionInfo`** for versioning
- **`PerformanceOptions`** to optimize delivery

### 🔹 PerformanceOptions & Compression

Electrician supports multiple compression algorithms for payload efficiency:

- **Zstd, Brotli, LZ4** – High-speed and optimized compression
- **Snappy, Deflate** – Balanced for speed and space efficiency
- **None** – No compression for raw binary transmission

### 🔹 RelayService (gRPC)

Defines RPC methods for **message transmission**:

- **`Receive`** – Accepts a single **WrappedPayload** and returns an acknowledgment
- **`StreamReceive`** – Enables **streaming** payloads between nodes or services in real-time

### 🔹 StreamAcknowledgment

Used to confirm successful processing:

- **`success`** flag (**true/false**)
- **`message`** field for additional status
- **`metadata`** for tracking response details

---

## 🚀 Usage

Electrician's protobuf definitions are designed for **cross-language compatibility** and can be compiled into various target languages (**Go, Python, Rust, etc.**) for seamless integration into different microservices and applications. These definitions allow:

- **Efficient serialization** of structured data
- **Secure and optimized communication** between Electrician components
- **Streaming & batch processing** via gRPC

### 🔹 Compiling the Protobuf Definitions

To generate client/server code for different languages, you need the `protoc` compiler and the appropriate plugin for your target language.

#### ✅ Generate **Go** Bindings:

```bash
protoc --go_out=. --go-grpc_out=. electrician.proto
```

🦀 Generate Rust Bindings (with tonic):

```bash
cargo install protobuf-codegen
cargo install tonic-build

# Generate Rust code
protoc --proto_path=. --rust_out=src/ --grpc_out=src/ --plugin=protoc-gen-grpc=`which grpc_rust_plugin` electrician.proto
```

📌 Note: Ensure tonic-build is added to your Cargo.toml dependencies.

🐍 Generate Python Bindings (with grpcio-tools):

```bash
pip install grpcio grpcio-tools

python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. electrician.proto
```

📌 Note: Python gRPC files will be generated, but you may need to manually adjust imports when using in a package.

📂 Organization
This protobuf definition is primarily used internally by Electrician’s relays, conduits, and streaming services.
Each Electrician package has its own dedicated README.md with further details.
The examples/ directory includes real-world implementations demonstrating how these messages are used in Electrician pipelines.

🤝 Contributing
If you have suggestions for improving the protobuf definitions—whether by adding new fields, supporting additional compression methods, or enhancing streaming capabilities—feel free to open an issue or submit a pull request.

📝 License
Electrician Protocol Buffers are released under the Apache 2.0 License.
You are free to use, modify, and distribute them under the terms of this license.

Happy wiring! ⚙️🚀 If you have any questions, feel free to open a GitHub issue or reach out.
