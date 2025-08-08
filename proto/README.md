# ⚡ Electrician Protocol Buffers (Protobuf)

The **Electrician Protocol Buffers** define the canonical wire format and gRPC service contract for all Electrician relays and pipelines.  
They are designed for **high-assurance, production-grade environments** where security, efficiency, and interoperability are mandatory.

Key design goals:

- **Type safety** — strict schema enforcement across all supported languages
- **Compression efficiency** — built-in support for multiple modern algorithms
- **Layer 7 encryption** — optional AES-GCM payload encryption on top of transport security
- **Integrated auth hints and context** — convey resource-server requirements and post-validation facts
- **Minimal coupling** — message format is stable and portable across deployments and org boundaries
- **Seamless data exchange** — optimized for local pipelines, distributed microservices, and cross-network delivery

Whether running in a single process, on a LAN, or across federated services with mTLS + OAuth2, these protobufs allow Electrician components to interoperate with **consistent structure and predictable semantics**.

## 📋 Protobuf Definitions

```protobuf
syntax = "proto3";

import "google/protobuf/timestamp.proto";

package electrician;

option go_package = "github.com/joeydtaylor/electrician/pkg/internal/relay";

// ---------- Core ----------

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
  repeated string details = 3;
}

message MessageMetadata {
  map<string, string> headers = 1;
  string content_type = 2;
  VersionInfo version = 3;
  PerformanceOptions performance = 4;
  string trace_id = 5;
  int32 priority = 6;

  // Existing content/transport crypto.
  SecurityOptions security = 7;

  // NEW (optional): auth hints for resource servers; server config may override.
  AuthenticationOptions authentication = 8;

  // NEW (optional): populated by receiver after auth; no raw tokens.
  AuthContext auth_context = 9;
}

message SecurityOptions {
  bool enabled = 1;
  EncryptionSuite suite = 2;
}

enum EncryptionSuite {
  ENCRYPTION_NONE = 0;
  ENCRYPTION_AES_GCM = 1;
}

message VersionInfo {
  int32 major = 1;
  int32 minor = 2;
}

message PerformanceOptions {
  bool use_compression = 1;
  CompressionAlgorithm compression_algorithm = 2;
  int32 compression_level = 3;
}

enum CompressionAlgorithm {
  COMPRESS_NONE = 0;
  COMPRESS_DEFLATE = 1;
  COMPRESS_SNAPPY = 2;
  COMPRESS_ZSTD = 3;
  COMPRESS_BROTLI = 4;
  COMPRESS_LZ4 = 5;
}

// ---------- Auth (optional) ----------

enum AuthMode {
  AUTH_MODE_UNSPECIFIED = 0;
  AUTH_NONE = 1;        // No app-layer auth (e.g., mTLS-only)
  AUTH_OAUTH2 = 2;      // Bearer token in "authorization" metadata
  AUTH_MUTUAL_TLS = 3;  // Client cert at transport layer
}

message AuthenticationOptions {
  bool enabled = 1;          // If false, treat as AUTH_NONE
  AuthMode mode = 2;         // Expected mode
  OAuth2Options oauth2 = 3;  // OAuth2/JWT/introspection (optional)
  MTLSOptions mtls = 4;      // mTLS expectations (optional)
}

message OAuth2Options {
  // Validation modes
  bool accept_jwt = 1;              // Validate JWT locally via JWKS
  bool accept_introspection = 2;    // RFC 7662 for opaque tokens

  // JWT validation hints (optional)
  string issuer = 3;                // Expected iss
  string jwks_uri = 4;              // Override JWKS URI; else discover
  repeated string required_audience = 5;
  repeated string required_scopes = 6;

  // Introspection config (optional)
  string introspection_url = 7;
  string introspection_auth_type = 8;   // "basic" | "bearer" | "none"
  string introspection_client_id = 9;
  string introspection_client_secret = 10;
  string introspection_bearer_token = 11;

  // Behavior (optional)
  bool forward_bearer_token = 12;       // Default false
  string forward_metadata_key = 13;     // e.g., "x-forwarded-authorization"

  // Caching hints (seconds, optional)
  int32 jwks_cache_seconds = 14;        // e.g., 300
  int32 introspection_cache_seconds = 15;// e.g., 60
}

message MTLSOptions {
  repeated string allowed_principals = 1; // CN/SAN/SPIFFE
  string trust_domain = 2;                // Impl-specific
}

// Auth facts emitted post-validation (no raw tokens).
message AuthContext {
  AuthMode mode = 1;
  bool authenticated = 2;
  string principal = 3;                 // Canonical subject
  string subject = 4;                   // Token sub or cert subject
  string client_id = 5;                 // OAuth2 client_id if present
  repeated string scopes = 6;
  map<string, string> claims = 7;       // Curated claim set (stringified)
  google.protobuf.Timestamp expires_at = 8;
  string issuer = 9;
  repeated string audience = 10;
  string token_id = 11;                 // JWT jti if present
}

// ---------- Service ----------

service RelayService {
  rpc Receive(WrappedPayload) returns (StreamAcknowledgment);
  rpc StreamReceive(stream WrappedPayload) returns (stream StreamAcknowledgment);
}

message StreamAcknowledgment {
  bool success = 1;
  string message = 2;
  map<string, string> metadata = 3;
}
```

## 📌 Message Structure & Components

### 🔹 WrappedPayload

Represents a structured data payload containing:

- A unique **`id`**
- A **`timestamp`** for event ordering
- The **`payload`** (application data, possibly compressed and/or encrypted)
- **`MessageMetadata`** for additional context (headers, tracing, priority, auth hints, etc.)
- **`ErrorInfo`** for error reporting

### 🔹 ErrorInfo

Encapsulates error details, including:

- A status **`code`**
- A descriptive **`message`**
- A **`details`** array for additional debugging information

### 🔹 MessageMetadata

Carries contextual metadata for each message:

- **`headers`** — custom key/value pairs
- **`content_type`** — MIME type of the payload
- **`version`** — structured major/minor version info
- **`performance`** — compression settings for the payload
- **`trace_id`** — distributed tracing identifier
- **`priority`** — integer priority indicator
- **`security`** — optional content-level encryption settings (e.g., AES-GCM)
- **`authentication`** — optional resource-server expectations for validating the sender
- **`auth_context`** — populated post-validation by the receiver; contains principal, scopes, claims, and token facts (never raw tokens)

### 🔹 AuthenticationOptions

Provides the receiver with the sender’s intended authentication mode and expectations:

- **Modes** — `AUTH_NONE`, `AUTH_OAUTH2`, or `AUTH_MUTUAL_TLS`
- **OAuth2Options** — JWKS and/or RFC 7662 introspection settings, required audiences/scopes, cache lifetimes
- **MTLSOptions** — allowed client principals and trust domain

This is **advisory metadata** — the receiver enforces its own configured policy and may ignore or override these hints.

### 🔹 AuthContext

Emitted by the receiver after successful authentication:

- **`mode`** — which auth mode was validated
- **`authenticated`** — boolean result
- **`principal`** — canonicalized subject
- **`subject`** — token subject or cert subject
- **`client_id`** — OAuth2 client identifier (if applicable)
- **`scopes`** — granted scopes
- **`claims`** — curated claim set
- **`expires_at`** — token/cert expiration
- **`issuer`** and **`audience`** — for auditing
- **`token_id`** — JWT `jti` if present

### 🔹 PerformanceOptions & Compression

Electrician supports multiple compression algorithms for payload efficiency:

- **Zstd, Brotli, LZ4** — high-speed, modern algorithms
- **Snappy, Deflate** — balanced for speed and space savings
- **None** — raw payload transmission

Optional **compression level** tuning is available per algorithm.

### 🔹 RelayService (gRPC)

Defines RPC methods for message delivery:

- **`Receive`** — sends a single `WrappedPayload` and receives an acknowledgment
- **`StreamReceive`** — streams multiple payloads bidirectionally for high-throughput or continuous data flows

### 🔹 StreamAcknowledgment

Acknowledges processing status:

- **`success`** — boolean indicator
- **`message`** — descriptive result
- **`metadata`** — additional structured key/value metadata

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

⚡ Regenerate Electrician Bindings (from project root):

```bash
protoc \
  -I=proto \
  --go_out=pkg/internal/relay \
  --go_opt=paths=source_relative \
  --go_opt=Mproto/electrician_relay.proto=pkg/internal/relay \
  --go-grpc_out=pkg/internal/relay \
  --go-grpc_opt=paths=source_relative \
  --go-grpc_opt=Mproto/electrician_relay.proto=pkg/internal/relay \
  proto/electrician_relay.proto
```

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
