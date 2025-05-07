# 🌐 HTTP Server

The **HTTP Server** exposes **webhooks or HTTP endpoints** within Electrician pipelines,  
allowing external services to **send data** into your pipeline and receive **custom responses**.

It supports **TLS, custom request parsing, flexible response handling, and structured logging**,  
making it ideal for **secure and straightforward inbound HTTP integrations**.

---

## 📦 Package Overview

| Feature                       | Description                                                               |
| ----------------------------- | ------------------------------------------------------------------------- |
| **TLS Support**              | Enables **secure HTTPS** endpoints.                                       |
| **Flexible Request Parsing**  | Decodes **JSON, XML, or raw bytes** into structured pipeline data.        |
| **Custom Response Handling** | Returns **custom HTTP status codes, headers, and JSON payloads**.         |
| **Lifecycle Logging**         | Provides **structured logs** for start-up, requests, errors, and more.    |
| **Sensor Observability**      | Hooks into **sensors** to track request lifecycle events (optional).      |

---

## 📂 Package Structure

| File            | Purpose                                                                 |
| --------------- | ----------------------------------------------------------------------- |
| **api.go**      | Public API for **configuring and managing HTTP server** instances.      |
| **internal.go** | Handles **request parsing**, TLS configuration, and **core server** logic. |
| **options.go**  | Functional options for **declarative HTTP server** setup.               |
| **httpserver.go** | Core **type definitions** and the **`NewHTTPServer`** constructor. |
| **README.md**   | This overview document describing the **HTTP Server**.          |

---

## 🔧 How the HTTP Server Works

The **HTTP Server** listens on a specified **address and endpoint** for incoming HTTP or HTTPS requests.  
It **decodes the request body** (JSON, XML, raw bytes, etc.) and **injects the data** into Electrician’s pipeline.  
You can **return custom responses**—including status codes, headers, and JSON payloads—back to the client.

### ✅ **Key Mechanisms**

- **Method & Path Matching:** Supports **specific endpoints** (e.g., `POST /webhook`) for inbound requests.
- **Custom Response Handling:** Easily define **status codes, headers, and JSON bodies**.
- **Security & TLS:** Optionally **enable HTTPS** with your certificate and key.
- **Sensor Integration:** Use **sensor hooks** to track request metrics, logs, and error events.
- **Timeout Management:** Safeguards against **long-running or hung connections**.

---

## 🔧 Extending the HTTP Server

To **add new features** or hooks, follow a structured approach:

1. **Enhance `types/`** – Add or update interfaces in `types/httpserver.go` to represent new behaviors.
2. **Extend `options.go`** – Introduce new **functional options** for configuration (e.g., request limit, advanced logging).
3. **Adapt `api.go`** – Expose updated functionality through the **public interface**.
4. **Modify `internal.go`** – Handle additional server logic, like **IP filtering**, advanced TLS, or custom routing.
5. **Refine `httpserver.go`** – Ensure the main server logic leverages any **new features** consistently.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Explore Electrician’s **overall architecture** and design principles.
- **[Examples Directory](../../../../example/httpserver)** – Demonstrates **HTTP Server** usage in a real-world scenario.

---

## 📝 License

The **HTTP Server** is part of Electrician, released under the [Apache 2.0 License](../../../LICENSE).  
Use, modify, and distribute it under these terms.

---

**Happy hosting!** If you have any questions or need guidance, open an issue on GitHub.
