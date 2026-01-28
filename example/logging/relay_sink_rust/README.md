# Rust Log Relay Sink (gRPC)

This README explains how a Rust service that already speaks the Electrician relay gRPC contract
should emit logs using the same log relay sink behavior as the Go implementation, or implement its
own sink with compatible wire format and semantics.

The contract lives in `proto/electrician_relay.proto` and the log schema is `electrician.log.v1`.
The Go sink implementation is in `pkg/internal/internallogger/relay_sink.go`.

## Decision: reuse vs. re-implement

- There is no shared Go sink library for Rust, so a Rust service should implement a native
  log relay sink that speaks the same gRPC contract.
- If you want parity with the Go sink, match the payload fields and metadata below. Your internal
  buffering, retries, and backpressure behavior can be different as long as you keep the wire
  contract intact.

## Wire contract (match the Go sink)

The Go sink streams `RelayEnvelope` messages to `RelayService.StreamReceive`:

- The stream starts with a `StreamOpen` message.
- Each log record is sent as a `WrappedPayload` inside a `RelayEnvelope`.
- The gRPC server returns `StreamAcknowledgment` messages; you should read them to avoid
  backpressure and to track failures.

`WrappedPayload` fields and metadata expected by the Go receiver:

- `payload`: JSON bytes representing a single log record (no trailing newline required).
- `metadata.content_type`: `application/json`
- `payload_type`: `electrician.log.v1`
- `payload_encoding`: leave as `PAYLOAD_ENCODING_UNSPECIFIED`
- `metadata.headers["source"]`: `electrician-logger`
- `metadata.headers["log_schema"]`: `electrician.log.v1`
- `timestamp`: set to now (UTC)
- `id`: RFC3339Nano string (used as a unique ID; any unique value is acceptable)
- `version`: `major=1`, `minor=0`

The Go sink does not set compression or encryption options for logs by default.

## Log schema (electrician.log.v1)

Log records are JSON objects with these reserved keys:

- `log_schema`: schema identifier (`electrician.log.v1`)
- `ts`: RFC3339Nano timestamp in UTC
- `level`: lowercase level string
- `msg`: log message
- `logger`, `caller`, `stack`: optional

Recommended fields for cross-team analytics:

- `component`: component metadata (type/id/name)
- `event`: event name
- `result`: result status
- `error`: error string
- `trace_id`, `span_id`: tracing IDs

Example payload:

```json
{"log_schema":"electrician.log.v1","ts":"2025-01-27T04:15:22.531Z","level":"info","msg":"relay log sample","event":"log_emit","seq":42,"component":{"type":"WIRE","id":"abc123"}}
```

## Recommended defaults (match Go sink)

The Go relay sink defaults are a good starting point for Rust:

- `queue_size`: 2048
- `submit_timeout`: 2s
- `flush_timeout`: 2s
- `drop_on_full`: true
- `auth_required`: true

If you configure bearer token auth, the Go sink refuses to dial without TLS. Mirror that behavior in Rust.

## Rust implementation outline

1. Generate Rust gRPC stubs from `proto/electrician_relay.proto` (e.g. using `tonic_build`).
2. Implement a log handler that turns a log record into a JSON object (schema above).
3. Open a bidirectional stream with `RelayService.StreamReceive`.
4. Send a `StreamOpen` envelope (recommended values below).
5. For each log record, send a `RelayEnvelope` containing a `WrappedPayload`.
6. Read `StreamAcknowledgment` responses in a background task.
7. On shutdown, flush your queue and send `StreamClose`.

Suggested `StreamOpen` values (match Go forwardrelay behavior):

- `ack_mode`: `ACK_BATCH`
- `ack_every_n`: 1024
- `max_in_flight`: 8192 (or your outbound queue size)
- `omit_payload_metadata`: true
- `defaults.headers["source"]`: `go` (not required if each payload sets metadata)
- `defaults.content_type`: `application/octet-stream` (unused if payload metadata is set)

## gRPC metadata and auth

The Go sink sets request metadata (gRPC headers) separately from payload metadata:

- `static_headers`: sent as gRPC metadata on the stream
- `authorization`: `Bearer <token>` if `bearer_token` or `bearer_token_env` is configured
- `trace-id`: set if not already present (any unique value is fine)

In Rust (tonic), add these to request metadata when opening the stream.

## Minimal Rust sketch (tonic + async)

```rust
use std::time::SystemTime;
use serde_json::json;
use tonic::metadata::MetadataValue;
use tonic::transport::Channel;
use tonic::Request;
use tokio::sync::mpsc;
use tokio_stream::wrappers::ReceiverStream;
use prost_types::Timestamp;

use electrician::relay_service_client::RelayServiceClient;
use electrician::{
    AckMode,
    MessageMetadata,
    RelayEnvelope,
    StreamClose,
    StreamOpen,
    VersionInfo,
    WrappedPayload,
};

const LOG_SCHEMA: &str = "electrician.log.v1";

fn now_ts() -> Timestamp {
    let now = SystemTime::now().duration_since(SystemTime::UNIX_EPOCH).unwrap();
    Timestamp {
        seconds: now.as_secs() as i64,
        nanos: now.subsec_nanos() as i32,
    }
}

fn make_payload(log_obj: serde_json::Value) -> WrappedPayload {
    let payload_bytes = serde_json::to_vec(&log_obj).unwrap();
    WrappedPayload {
        id: chrono::Utc::now().to_rfc3339_opts(chrono::SecondsFormat::Nanos, true),
        timestamp: Some(now_ts()),
        payload: payload_bytes,
        payload_type: LOG_SCHEMA.to_string(),
        payload_encoding: 0, // PAYLOAD_ENCODING_UNSPECIFIED
        metadata: Some(MessageMetadata {
            headers: [
                ("source".to_string(), "electrician-logger".to_string()),
                ("log_schema".to_string(), LOG_SCHEMA.to_string()),
            ]
            .into_iter()
            .collect(),
            content_type: "application/json".to_string(),
            version: Some(VersionInfo { major: 1, minor: 0 }),
            ..Default::default()
        }),
        ..Default::default()
    }
}

async fn stream_logs(target: String) -> Result<(), Box<dyn std::error::Error>> {
    let channel = Channel::from_shared(format!("https://{}", target))?.connect().await?;
    let mut client = RelayServiceClient::new(channel);

    let (tx, rx) = mpsc::channel(1024);
    let mut request = Request::new(ReceiverStream::new(rx));

    // Add gRPC metadata headers.
    request.metadata_mut().insert("trace-id", MetadataValue::from_static("rust-relay"));

    tokio::spawn(async move {
        let open = StreamOpen {
            stream_id: "rust-log-relay".to_string(),
            ack_mode: AckMode::AckBatch as i32,
            ack_every_n: 1024,
            max_in_flight: 8192,
            omit_payload_metadata: true,
            ..Default::default()
        };
        tx.send(RelayEnvelope { msg: Some(electrician::relay_envelope::Msg::Open(open)) }).await.ok();

        let log_obj = json!({
            "log_schema": LOG_SCHEMA,
            "ts": "2025-01-27T04:15:22.531Z",
            "level": "info",
            "msg": "relay log sample",
            "event": "log_emit",
            "seq": 1
        });
        let payload = make_payload(log_obj);
        tx.send(RelayEnvelope { msg: Some(electrician::relay_envelope::Msg::Payload(payload)) }).await.ok();

        let close = StreamClose { reason: "shutdown".to_string() };
        tx.send(RelayEnvelope { msg: Some(electrician::relay_envelope::Msg::Close(close)) }).await.ok();
    });

    // Consume acks to avoid backpressure.
    let mut resp = client.stream_receive(request).await?.into_inner();
    while let Some(_ack) = resp.message().await? {}

    Ok(())
}
```

Notes:

- The snippet omits TLS client certs and bearer token setup; add them if required.
- Use a real unique ID for `WrappedPayload.id` (UUID or RFC3339Nano timestamp).
- In production, use a queue and a background sender to avoid blocking the log path.

## Test against the Go receiver

Start the Go receiver from this repo:

```bash
go run example/logging/relay_sink_receiver/main.go
```

Then point your Rust sink at `localhost:50090` and verify the logs are captured and decoded
with the expected schema.

## References

- `proto/electrician_relay.proto`
- `pkg/internal/internallogger/README.md`
- `example/logging/relay_sink_receiver/README.md`
