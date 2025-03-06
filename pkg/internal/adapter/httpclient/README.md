# 🌐 HTTP Client Adapter

The **HTTP Client Adapter** integrates **external data sources** into Electrician pipelines,  
fetching and transforming **HTTP responses** into structured pipeline data.

It supports **OAuth2, TLS pinning, interval-based polling, retries, and structured logging**,  
making it ideal for **secure and resilient HTTP integrations**.

---

## 📦 Package Overview

| Feature                   | Description                                                         |
| ------------------------- | ------------------------------------------------------------------- |
| **OAuth2 Authentication** | Supports **client credentials flow** for secure API access.         |
| **TLS Pinning**           | Enforces **certificate verification** for enhanced security.        |
| **Retry & Backoff**       | Implements **exponential backoff** for handling transient failures. |
| **Custom Headers**        | Supports **dynamic headers** (e.g., API keys, tokens, user-agents). |
| **JSON, XML, and Binary** | Decodes **structured and unstructured** HTTP responses.             |

---

## 📂 Package Structure

| File                   | Purpose                                                          |
| ---------------------- | ---------------------------------------------------------------- |
| **api.go**             | Public API for **configuring and managing HTTP clients**.        |
| **internal.go**        | Handles **response parsing, TLS verification, and retry logic**. |
| **notify.go**          | Structured logging and **sensor-based event tracking**.          |
| **options.go**         | Functional options for **configurable HTTP request behavior**.   |
| **httpclient.go**      | Core **Type Definition and Constructor**.                        |
| **httpclient_test.go** | Unit tests ensuring **error handling, retries, and security**.   |

---

## 🔧 How the HTTP Client Adapter Works

The **HTTP Client Adapter** **fetches data from APIs** and transforms it into structured pipeline output.

### ✅ **Key Mechanisms**

- **Interval-Based Requests:** Periodically fetches **HTTP data** with retry logic.
- **OAuth2 Authentication:** **Requests and refreshes access tokens** dynamically.
- **TLS Certificate Pinning:** Ensures **secure HTTPS connections**.
- **Custom Headers & Metadata:** Configurable **per-request headers and parameters**.
- **Multi-Format Decoding:** Handles **JSON, XML, binary, and plain text responses**.

---

## 🔧 Extending the HTTP Client Adapter

To **add new features**, follow this **structured workflow**:

### 1️⃣ Modify `types/`

- Define new methods inside `types/httpclient.go`.
- This ensures **consistent interface support** across Electrician.

### 2️⃣ Implement in `api.go`

- Update the **public API** to expose new configurations.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **declarative HTTP client configuration**.

### 4️⃣ Extend `notify.go` for logging & telemetry

- If new events are introduced, add **sensor and logger hooks**.

### 5️⃣ Unit Testing (`httpclient_test.go`)

- **Validate authentication, retries, and response handling**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Examples Directory](../../../../example/plug_example/httpadapter/)** – Demonstrates **HTTP Client Adapter in action**.

---

## 📝 License

The **HTTP Client Adapter** is part of Electrician and is released under the [MIT License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy fetching! 🌍🔗** If you have questions or need support, feel free to open a GitHub issue.
