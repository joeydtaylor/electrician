# 🌐 HTTP Server Package

Electrician’s HTTP server is a **small ingress wrapper** around Go’s `net/http`.

It’s used to expose a **webhook-style endpoint** that:

* accepts inbound HTTP requests
* reads/decodes the request body into a value of type `T`
* submits that value into your pipeline (typically a `Wire` or other `types.Submitter[T]`)
* returns a response (status/headers/body) based on your configured behavior

It’s not a general-purpose web framework. It’s a pipeline ingress.

---

## ✅ What it does

| Capability            | Meaning                                                                     |
| --------------------- | --------------------------------------------------------------------------- |
| 🔌 Inbound ingestion  | Turn HTTP requests into pipeline submissions.                               |
| 🧩 Pluggable decoding | You control how request bodies become `T` (raw bytes, JSON, custom).        |
| ↩️ Response control   | You can decide what gets returned (ack-style responses, errors, headers).   |
| 🔐 TLS support        | Uses standard library TLS (`http.Server` + `tls.Config`) when configured.   |
| ⏱️ Timeouts + limits  | Server-level timeouts and request-size limits to avoid hangs/abuse.         |
| 📡 Telemetry hooks    | Optional log/sensor hooks around request lifecycle (implementation-driven). |

---

## 🧠 How it fits in a pipeline

Common shape:

**HTTP Server → Wire (or Submitter) → downstream pipeline**

The HTTP server should do as little work as possible:

* validate headers
* decode body → `T`
* submit
* return

Heavy work belongs in the pipeline, not in the HTTP handler.

---

## 🔧 Request handling model

A typical request path looks like:

1. **Match** method + path (webhook endpoint)
2. **Read** the request body (bounded)
3. **Decode** into `T` (codec/decoder you configured)
4. **Submit** into the configured downstream component
5. **Respond** with a configured status/body

Notes:

* “JSON/XML support” is not magic. If you want JSON decoding, wire in a JSON decoder. Same for anything else.
* Submission is only as reliable as your downstream pipeline. If you need durable delivery, use a broker upstream.

---

## 🔐 TLS + auth posture

* TLS is transport security. This package can run HTTPS when you provide TLS config/credentials.
* Authentication/authorization is a policy decision. If you need auth, enforce it explicitly (mTLS, JWT validation, header checks, IP allowlists, etc.).

Do not assume “secure by default” beyond what you configure.

---

## 📂 Package structure

| File            | Purpose                                 |
| --------------- | --------------------------------------- |
| `httpserver.go` | Type definition + constructor           |
| `api.go`        | Public methods / configuration wiring   |
| `internal.go`   | Handler + server implementation details |
| `options.go`    | Functional options for configuration    |
| `*_test.go`     | Tests                                   |

---

## 🧰 Extension workflow

When adding capability, keep layering clean:

* Cross-component contract → update `types/httpserver.go` first.
* Implementation details → `pkg/internal/httpserver`.
* User-facing knobs → expose through `pkg/builder` (`HTTPServerWith…`).

Tests should cover:

* request decode success/failure
* response shaping
* context cancellation + timeouts
* safe shutdown
* limits (max body, etc.)

## 📖 Further Reading

- **[Root README](../../../README.md)** – Explore Electrician’s **overall architecture** and design principles.
- **[Examples Directory](../../../../example/httpserver)** – Demonstrates **HTTP Server** usage in a real-world scenario.

---

## 📝 License

The **HTTP Server** is part of Electrician, released under the [Apache 2.0 License](../../../LICENSE).  
Use, modify, and distribute it under these terms.

---

**Happy hosting!** If you have any questions or need guidance, open an issue on GitHub.
