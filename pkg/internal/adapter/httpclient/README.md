# 🌐 HTTP Client Adapter

The **HTTP Client Adapter** is an ingestion adapter that turns **HTTP responses** into typed values (`T`) for Electrician pipelines.

It’s typically used behind a `Plug` and driven by a `Generator` (polling) or any other component that wants to fetch remote data and feed it into a wire.

---

## ✅ What it does

| Capability                       | Meaning                                                                                                         |
| -------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| 🌍 HTTP fetch                    | Make requests with `net/http` using a configured method/URL/headers/body.                                       |
| 🧬 Typed decode                  | Decode the response body into `T` using a configured decoder/codec.                                             |
| ⏱️ Timeouts                      | Bound request lifetime (client/server timeouts) and respect context cancellation.                               |
| 📡 Telemetry hooks               | Optional sensor/logger events around request/response/errors (implementation-driven).                           |
| 🔁 Polling support (when driven) | In polling setups, an interval is used to control cadence (usually via the generator or adapter configuration). |

---

## ❌ What it does *not* do by default

* ❌ It is not a full HTTP framework.
* ❌ It is not durable delivery.
* ❌ It does not automatically give you OAuth2, JWT validation, or TLS pinning unless you explicitly wire those behaviors.

Electrician’s posture is explicit wiring: if you need auth, retries, or custom TLS behavior, you configure it or wrap it.

---

## 🧠 How it fits in a pipeline

Common composition:

**HTTP Client Adapter → Plug → Generator → Wire**

* Adapter: owns HTTP request + decoding.
* Plug: owns adapter wiring and optional ingestion shaping.
* Generator: decides when to fetch (poll cadence / triggering).
* Wire: does concurrent transformation and output.

Keep the adapter focused on IO + decode. Put business logic in wire transformers.

---

## 🔧 Request model

At minimum, an HTTP adapter typically needs:

* method (`GET`, `POST`, …)
* URL
* headers
* optional body
* timeout settings

Context cancellation should cancel in-flight requests cleanly.

If you need custom transport behavior (proxies, custom root CAs, mTLS, pinning), use a custom `http.Client` / `http.Transport` configuration where the adapter exposes that hook.

---

## 🧾 Response decoding

Decoding should be treated as **pluggable**:

* JSON: `encoding/json`
* XML: `encoding/xml`
* raw bytes / text: `io` + `bufio`
* custom: your own decoder

The adapter should not guess formats. If you want JSON, wire a JSON decoder. If the remote returns bytes, decode bytes.

---

## 🪪 Authentication posture

Auth is usually handled by:

* setting `Authorization`/API-key headers, or
* wrapping the adapter with a token provider that refreshes and injects headers.

Do not assume the adapter provides a full OAuth2 implementation unless you can point to the concrete option/implementation in code.

---

## 📂 Package structure

| File            | Purpose                              |
| --------------- | ------------------------------------ |
| `httpclient.go` | Type definition + constructor        |
| `api.go`        | Public methods / wiring              |
| `internal.go`   | Request execution + decode logic     |
| `options.go`    | Functional options for configuration |
| `*_test.go`     | Tests                                |

---

## 🔧 Extending the adapter

* Contract changes start in `pkg/internal/types/httpclient_adapter.go`.
* Implementation changes go in this package.
* User-facing knobs should be exposed via `pkg/builder` (`HTTPClientAdapterWith…`).

Tests should cover:

* request building + header behavior
* decode success/failure
* timeouts + cancellation
* error classification (transport vs decode)

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Examples Directory](../../../../example/plug_example/httpadapter/)** – Demonstrates **HTTP Client Adapter in action**.

---

## 📝 License

The **HTTP Client Adapter** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy fetching! 🌍🔗** If you have questions or need support, feel free to open a GitHub issue.
